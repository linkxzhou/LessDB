package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/linkxzhou/LessDB/internal/utils"

	"github.com/labstack/echo/v4"

	_ "github.com/mattn/go-sqlite3"
)

type (
	ExecuteDBParams struct {
		QueryParams
		CMDListParams
		WriteKey string `json:"writekey" form:"writekey" query:"writekey" param:"writekey"`
	}

	ExecuteLogParams struct {
		QueryParams
		SeqID string `json:"seqid" form:"seqid" query:"seqid" param:"seqid"`
	}

	UploadS3Redolog struct {
		DBName        string                  `json:"dbname"`
		ReadKey       string                  `json:"readkey"`
		WriteKey      string                  `json:"writekey"`
		NanoTimestamp int64                   `json:"nanotimestamp"`
		List          []SQLExecuteCommandArgs `json:"list"`
	}

	ResultRedolog struct {
		SeqID   string `json:"seqid"`
		Message string `json:"message"`
		Status  int    `json:"status"`
	}
)

const (
	ErrNoAuthWrite = "attempt to write a readonly database"
	// Execute sql has status on sqlite3
	ExecStatusPending int = iota
	ExecStatusOK
	ExecStatusFailed
	ExecStatusPartialFailed
)

// ExecuteDB for exec sql on sqlite3 file
func ExecuteDB(c echo.Context) error {
	edbp := new(ExecuteDBParams)
	if err := c.Bind(edbp); err != nil || edbp.List == nil || len(edbp.List) < 0 {
		return c.String(http.StatusBadRequest, "bad request")
	}

	c.Logger().Info("ExecuteDB edbp: ", edbp)

	readKey := edbp.ReadKey
	dbName, authOK := utils.VerifyName(readKey)
	if !authOK || dbName == "" {
		return c.JSON(http.StatusOK, DataResp{
			Code:    -999,
			Message: "No Auth",
		})
	}

	db, uri, err := getVFSDB(dbName)
	if err != nil {
		c.Logger().Error("getVFSDB err: ", err)
		return err
	}
	defer db.Close()

	c.Logger().Info("S3 GetFileLink: ", uri, ", dbName: ", dbName)
	var redologs []SQLExecuteCommandArgs
	for _, cmd := range edbp.List {
		err := execSQLWithHTTPVFS(c, db, cmd)
		if err != nil {
			if !strings.Contains(err.Error(), ErrNoAuthWrite) {
				c.Logger().Error("ExecuteSQL err: ", err)
				return err
			}
			redologs = append(redologs, cmd)
		}
	}

	// Upload redolog to S3
	nanoTimestamp := time.Now().UnixNano()
	uploadS3Redolog := UploadS3Redolog{
		DBName:        dbName,
		NanoTimestamp: nanoTimestamp,
		ReadKey:       readKey,
		WriteKey:      edbp.WriteKey,
		List:          redologs,
	}
	jsons, err := json.Marshal(uploadS3Redolog)
	if err != nil {
		c.Logger().Error("Marshal err: ", err)
		return err
	}

	s3key := fmt.Sprintf("%v-%v.redolog", readKey, nanoTimestamp)
	err = s3Client.UploadString(context.TODO(), s3key, jsons)
	if err != nil {
		c.Logger().Error("S3 UploadFile err: ", err)
		return err
	}

	return c.JSON(http.StatusOK, DataResp{
		Code:    0,
		Message: "OK",
		Data: ResultRedolog{
			SeqID:   s3key,
			Message: "Execute Pending",
			Status:  ExecStatusPending,
		},
	})
}

// ExecuteLog for query redolog on sqlite3 file
func ExecuteLog(c echo.Context) error {
	elp := new(ExecuteLogParams)
	if err := c.Bind(elp); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	c.Logger().Info("ExecuteLogParams elp: ", elp)

	readKey := elp.ReadKey
	dbName, authOK := utils.VerifyName(readKey)
	if !authOK || dbName == "" {
		return c.JSON(http.StatusOK, DataResp{
			Code:    -999,
			Message: "No Auth",
		})
	}

	db, uri, err := getVFSDB(dbName)
	if err != nil {
		c.Logger().Error("getVFSDB err: ", err)
		return err
	}
	defer db.Close()

	c.Logger().Info("S3 GetFileLink: ", uri, ", dbName: ", dbName)
	_, values, _, _, err := querySQLWithHTTPVFS(c, db,
		SQLExecuteCommandArgs{
			CMD:  fmt.Sprintf(`SELECT name, value, value_int FROM %v where name = ?`, systemDBName),
			Args: []interface{}{elp.SeqID},
		})
	if err != nil {
		c.Logger().Error("ExecuteSQL err: ", err)
		return err
	}

	// TODO: fix values to redolog

	return c.JSON(http.StatusOK, DataResp{
		Code:    0,
		Message: "OK",
		Data:    values,
	})
}

// QueryDB for query data on sqlite3 file
func QueryDB(c echo.Context) error {
	edbp := new(ExecuteDBParams)
	if err := c.Bind(edbp); err != nil || edbp.List == nil || len(edbp.List) < 0 {
		return c.String(http.StatusBadRequest, "bad request")
	}

	c.Logger().Info("QueryDB edbp: ", edbp)

	var result []DBValuesResp
	readKey := edbp.ReadKey
	dbName, authOK := utils.VerifyName(readKey)
	if !authOK || dbName == "" {
		return c.JSON(http.StatusOK, DataResp{
			Code:    -999,
			Message: "No Auth",
		})
	}

	db, uri, err := getVFSDB(dbName)
	if err != nil {
		c.Logger().Error("getVFSDB err: ", err)
		return err
	}
	defer db.Close()

	c.Logger().Info("S3 GetFileLink: ", uri, ", dbName: ", dbName)
	for _, cmd := range edbp.List {
		columns, values, types, count, err := querySQLWithHTTPVFS(c, db, cmd)
		if err != nil {
			c.Logger().Error("ExecuteSQL err: ", err)
			return err
		}
		result = append(result, DBValuesResp{
			Columns: columns,
			Values:  values,
			Types:   types,
			Count:   count,
		})
	}

	return c.JSON(http.StatusOK, DataResp{
		Code:    0,
		Message: "OK",
		Data:    result,
	})
}
