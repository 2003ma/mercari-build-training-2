package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"

	"mime/multipart"
)

const (
	ImgDir = "images"
)

const (
	DbDir="../../db"
)

type Response struct {
	Message string `json:"message"`
}

func errMessage(c echo.Context, err error, status int, message string) error {
	errorMessage := fmt.Sprintf("%s:%s", message, err)
	return c.JSON(status, Response{Message: errorMessage})
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

func hashImg(c echo.Context, image *multipart.FileHeader)(string,error){
	hash:=sha256.New()

	imgFile,err:=image.Open()
	if err != nil {
		errMessage(c, err, http.StatusBadRequest, "Unable to open the image")
	}
	defer imgFile.Close()

	if _, err := io.Copy(hash, imgFile); err != nil {
		errMessage(c, err, http.StatusBadRequest, "Unable to copy imgFile to hash")
	}

	hashValue := hash.Sum(nil)
	hashString := hex.EncodeToString(hashValue)
	imgName := hashString + ".jpg"

	return imgName,err
}

func addItem(c echo.Context) error {
	name := c.FormValue("name")
	c.Logger().Infof("Receive item: %s", name)

	message := fmt.Sprintf("item received: %s", name)
	res := Response{Message: message}

	
	return c.JSON(http.StatusOK, res)
}


func addCategory(c echo.Context) error{
	category := c.FormValue("category")
	message := fmt.Sprintf("category received: %s",category)
	res := Response{Message: message}

	db,err:=sql.Open("sqlite3",DbDir)
	if err != nil {
		return errMessage(c, err, http.StatusBadRequest, "Unable to open database")
	}
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO categories(name) VALUES (?)")
	if err != nil {
		return errMessage(c, err, http.StatusBadRequest, "Unable to open database")
	}
	defer stmt.Close()

	_, err = stmt.Exec(category)
	if err != nil {
		return errMessage(c, err, http.StatusBadRequest, "Unable to execute sql command")
	}

	return c.JSON(http.StatusOK, res)
}


func getImg(c echo.Context) error {
	imgPath := path.Join(ImgDir, c.Param("imageFilename"))

	if !strings.HasSuffix(imgPath, ".jpg") {
		res := Response{Message: "Image path does not end with .jpg"}
		return c.JSON(http.StatusBadRequest, res)
	}
	if _, err := os.Stat(imgPath); err != nil {
		c.Logger().Debugf("Image not found: %s", imgPath)
		imgPath = path.Join(ImgDir, "default.jpg")
	}
	return c.File(imgPath)
}

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Logger.SetLevel(log.INFO)

	frontURL := os.Getenv("FRONT_URL")
	if frontURL == "" {
		frontURL = "http://localhost:3000"
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{frontURL},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	// Routes
	e.GET("/", root)
	e.POST("/items", addItem)
	e.POST("/category",addCategory)
	e.GET("/image/:imageFilename", getImg)


	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
