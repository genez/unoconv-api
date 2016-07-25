package main

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/fasthttp"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"os/exec"
	"unoconv-api/unoconv"
)


func main() {

	uno := new(unoconv.UnoConv)
	uno.RequestChan = make(chan unoconv.Request)

	//unoconv can only process one file at a time
	go func(uno *unoconv.UnoConv) {
		for {
			select {
			case data := <-uno.RequestChan:
				cmd := exec.Command("unoconv", "-f", data.Filetype, "--stdout", data.Filename)
				cmd.Stdout = data.W
				err := cmd.Run()
				if err != nil {
					data.ErrChan <- err
				} else {
					data.ErrChan <- nil
				}
			}
		}
	}(uno)

	e := echo.New()
	e.GET("/unoconv/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "200 -OK")
	})
	e.POST("/unoconv/:filetype", func(c echo.Context) (err error) {
		file, err := c.FormFile("file")
		if err != nil {
			log.Print(err)
			return
		}
		src, err := file.Open()
	    if err != nil {
		return err
	    }
	    defer src.Close()

		//create a temporary file and copy the file from the form to it
		tempfile, err := ioutil.TempFile(os.TempDir(), "unoconv-api")
		if err != nil {
			log.Print(err)
			return
		}
		io.Copy(tempfile, src)
		tempfile.Close()

		//append the file extension to the temporary file's name
		filename := tempfile.Name() + filepath.Ext(file.Filename)
		os.Rename(tempfile.Name(), filename)
		defer os.Remove(filename)

		//Run unoconv to convert the file
		//unoconv's stdout is plugged directly to the httpResponseWriter
		err = uno.Convert(filename, c.Param("filetype"), c.Response().Writer())
		if err != nil {
			log.Print(err)
			return
		}
		return c.NoContent(http.StatusOK)
	})
	e.Run(fasthttp.New(":3000"))
}
