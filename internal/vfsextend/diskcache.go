package vfsextend

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/linkxzhou/LessDB/internal/utils"
)

const DefaultPageSize = 1 << 10
const DefaultNoCacheSize = -1

var envPageSize = utils.GetEnviron("CACHE_PAGESIZE")

type (
	DiskCacheHandler struct {
		PageSize int

		fileHandler FileHandler

		fileName string
		fileSize int64

		f     *os.File
		pages sync.Map

		cacheEtag string
		cacheHit  int
		cacheMiss int

		rwMutex sync.RWMutex
	}

	FileHandler func(string) (*os.File, error)
)

func NewDiskCache(fileHandler FileHandler, fileSize int64) *DiskCacheHandler {
	h := &DiskCacheHandler{
		fileHandler: fileHandler,
		fileSize:    fileSize,
	}

	if h.PageSize < 1 {
		// Set page size
		h.PageSize = DefaultPageSize
		pageSize, err := strconv.ParseInt(envPageSize, 10, 64)
		if err == nil {
			h.PageSize = int(pageSize)
		}
	}

	return h
}

func (h *DiskCacheHandler) newFileCache(etag string) (err error) {
	h.rwMutex.Lock()
	defer h.rwMutex.Unlock()

	if h.f == nil && h.fileHandler != nil {
		h.f, err = h.fileHandler(fmt.Sprintf("%v.%v", h.fileName, etag))
		if err == nil {
			h.cacheEtag = etag
			h.pages = *new(sync.Map)
		}
	}

	return err
}

func (h *DiskCacheHandler) Get(p []byte, off int64, fetcher VFSReadAt) (int, error) {
	// Cache filesize
	if h.fileSize <= DefaultNoCacheSize {
		if _, sizeErr := h.Size(fetcher); sizeErr != nil {
			return 0, sizeErr
		}
	}

	// Create new cache
	etag := fetcher.Etag()
	if h.EtagModify(etag) || h.f == nil {
		if err := h.newFileCache(etag); err != nil {
			return 0, err
		}
	}

	startPage, endPage := h.pagesForRange(off, len(p))

	firstMissingPage := int64(-1)
	lastMissingPage := int64(-1)
	for i := int64(startPage); i <= endPage; i++ {
		if v, ok := h.pages.Load(i); ok && v.(bool) {
			continue
		}
		if firstMissingPage < 0 {
			firstMissingPage = i
		}
		if lastMissingPage < i {
			lastMissingPage = i
		}
	}

	lastPage := h.fileSize / int64(h.PageSize)
	if firstMissingPage >= 0 {
		h.cacheMiss++
		pageCount := (lastMissingPage + 1) - firstMissingPage
		size := pageCount * int64(h.PageSize)
		if lastMissingPage == lastPage {
			size = size - int64(h.PageSize) + (h.fileSize % int64(h.PageSize))
		}
		offset := firstMissingPage * int64(h.PageSize)
		if h.fileSize >= 0 && h.fileSize < (size+offset) {
			size = h.fileSize - offset
		}
		buffer := make([]byte, size)
		n, readAtErr := fetcher.ReadAt(buffer, offset)
		if readAtErr != nil {
			return 0, readAtErr
		}

		// ReadAt has modify etag
		etag = fetcher.Etag()
		if h.EtagModify(etag) {
			h.cacheEtag = etag
			h.fileSize = DefaultNoCacheSize
			return 0, missCacheErr
		}

		buffer = buffer[:n]
		if wn, writeErr := h.f.WriteAt(buffer, offset); writeErr != nil && wn != n {
			return 0, writeErr
		}
		fullPagesRead := n / h.PageSize
		for i := int64(0); i < int64(fullPagesRead); i++ {
			h.pages.Store(firstMissingPage+i, true)
		}
	} else {
		h.cacheHit++
	}

	return h.f.ReadAt(p, off)
}

func (h *DiskCacheHandler) Size(fetcher VFSReadAt) (int64, error) {
	if h.fileSize <= DefaultNoCacheSize {
		fileSize, sizeErr := fetcher.Size()
		if sizeErr != nil {
			return 0, sizeErr
		}
		h.fileSize = fileSize
	}

	return h.fileSize, nil
}

func (h *DiskCacheHandler) SetFileName(name string) {
	h.fileName = name
}

func (h *DiskCacheHandler) EtagModify(etag string) bool {
	return h.cacheEtag != etag
}

func (h *DiskCacheHandler) pagesForRange(offset int64, size int) (startPage, endPage int64) {
	startPage = offset / int64(h.PageSize)
	endPage = (offset+int64(size))/int64(h.PageSize) + 1
	return startPage, endPage
}
