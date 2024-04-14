package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/linkxzhou/LessDB/internal/s3"
)

type (
	TiggerReq struct {
		Events []TiggerEvents `json:"events"`
	}

	TiggerEvents struct {
		S3Bucket string `json:"s3bucket"`
		S3Key    string `json:"s3key"`
	}
)

func TiggerS3Events(c echo.Context) error {
	req := new(TiggerReq)
	if err := c.Bind(req); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	for _, event := range req.Events {
		s3Client := s3.NewS3Client(S3Endpoint, S3Region, S3AccessKey, S3SecretKey, event.S3Bucket)
		if s3Client == nil {
			c.Logger().Error("NewS3Client failed!")
			return c.String(http.StatusBadRequest, "bad request")
		}

		redologStr, err := s3Client.DownloadString(context.TODO(), event.S3Key)
		if err != nil {
			c.Logger().Error("Download err: ", err)
			return err
		}

		var redolog UploadS3Redolog
		err = json.Unmarshal(redologStr, &redolog)
		if err != nil {
			c.Logger().Error("Unmarshal err: ", err)
			return err
		}

		dbName := redolog.DBName
		dbFile, err := os.Create(dbName)
		if err != nil {
			c.Logger().Error("OpenFile err: ", err)
			return err
		}
		defer func() {
			dbFile.Close()
			// Remove tempfile
			os.Remove(dbName)
		}()

		err = s3Client.Download(context.TODO(), dbName, dbFile)
		if err != nil {
			c.Logger().Error("Download err: ", err)
			return err
		}

		db, err := GetFileDB(dbName)
		if err != nil {
			c.Logger().Error("sql.Open err: ", err)
			return err
		}
		defer db.Close()

		// Insert into redolog result to system table
		err = insertSYSTEMDBStatus(c, db, ExecStatusPending, event.S3Key, "Pending")
		if err != nil {
			c.Logger().Error("insertSYSTEMDBStatus err: ", err)
			return err
		}

		// Execute redolog list on sqlite3 file
		err = ExecuteSQLWithFile(c, db, redolog.List)
		if err != nil {
			c.Logger().Error("ExecuteSQL err: ", err)
			// Update redolog error to system table
			err = updateSYSTEMDBStatus(c, db, ExecStatusFailed, event.S3Key, err.Error())
			if err != nil {
				c.Logger().Error("updateSYSTEMDBStatus err: ", err)
				return err
			}
		} else {
			// Update redolog ok to system table
			err = updateSYSTEMDBStatus(c, db, ExecStatusOK, event.S3Key, "OK")
			if err != nil {
				c.Logger().Error("updateSYSTEMDBStatus err: ", err)
				return err
			}
		}

		dbFile.Seek(0, io.SeekStart)
		c.Logger().Info("s3Client Upload: ", s3Client.String(dbName))
		err = s3Client.Upload(context.TODO(), dbName, dbFile)
		if err != nil {
			c.Logger().Error("S3 UploadFile err: ", err)
			return err
		}
	}

	return c.JSON(http.StatusOK, DataResp{
		Code:    0,
		Message: "OK",
	})
}

func insertSYSTEMDBStatus(c echo.Context, db *sql.DB, status int, name, value string) error {
	return ExecuteSQLWithFile(c, db, []SQLExecuteCommandArgs{
		SQLExecuteCommandArgs{
			CMD:  fmt.Sprintf(`INSERT INTO %v(value_int, value, name) values(?, ?, ?)`, systemDBName),
			Args: []interface{}{status, value, name},
		},
	})
}

func updateSYSTEMDBStatus(c echo.Context, db *sql.DB, status int, name, value string) error {
	return ExecuteSQLWithFile(c, db, []SQLExecuteCommandArgs{
		SQLExecuteCommandArgs{
			CMD:  fmt.Sprintf(`UPDATE %v SET value_int = ?, value = ? where name = ?`, systemDBName),
			Args: []interface{}{status, value, name},
		},
	})
}
