package remote

import (
	"callumj.com/weave/remote/s3"
	"callumj.com/weave/remote/uptypes"
	"callumj.com/weave/tools"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

type DownloadInfo struct {
	FilePath string
	ETag     string
}

func UploadToS3(config uptypes.S3Config, files []uptypes.FileDescriptor) {
	s3.UploadToS3(config, files)
}

func DownloadRemoteFile(url, finalDirectory string) *DownloadInfo {

	etagFile := fmt.Sprintf("%v/%v", finalDirectory, ".weave.etag")

	var eTag string
	if tools.PathExists(etagFile) {
		contents, err := ioutil.ReadFile(etagFile)
		if err != nil {
			log.Printf("Unable to read eTag file %v\r\n", etagFile)
		}
		eTag = string(contents)
	} else {
		eTag = ""
	}

	out, err := ioutil.TempFile("/tmp", "weave")
	if err != nil {
		log.Printf("Unable to create temp file\r\n")
		return nil
	}
	defer out.Close()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Unable to construct URL for %v\r\n", url)
		return nil
	}
	if len(eTag) != 0 {
		req.Header.Add("If-None-Match", eTag)
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Unable to communicate with server %v\r\n", url)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == 304 {
		log.Printf("Object not modified, finishing up.\r\n")
		return nil
	}

	n, err := io.Copy(out, resp.Body)
	if err != nil {
		log.Printf("Unable to download file\r\n")
		return nil
	}

	if n == 0 {
		log.Printf("Nothing was copied\r\n")
		return nil
	}

	return &DownloadInfo{
		FilePath: out.Name(),
		ETag:     resp.Header.Get("ETag"),
	}
}
