package handler

import (
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/labstack/echo/v4"
	"github.com/malikfajr/beli-mang/internal/entity/converter"
	"github.com/malikfajr/beli-mang/internal/exception"
	"github.com/oklog/ulid/v2"
)

var (
	AWS_S3_BUCKET_NAME = os.Getenv("AWS_S3_BUCKET_NAME")
	AWS_REGION         = os.Getenv("AWS_REGION")
)

type ImageHandler struct {
}

func (i *ImageHandler) Upload(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(400, exception.BadRequest("file required"))
	}

	// Check file size (2MB = 2 * 1024 * 1024 bytes)
	if file.Size > 2*1024*1024 || file.Size < 10*1024 {
		return c.JSON(400, exception.BadRequest("File size min 10kb, max 2MB limit"))
	}

	// split file name
	ext := strings.Split(file.Filename, ".")

	// create new file name
	key := ulid.Make().String() + "." + ext[len(ext)-1]

	// open file upload
	src, err := file.Open()
	if err != nil {
		panic(err)
	}
	defer src.Close()

	// check file
	if _, err := i.isValidImage(src); err != nil {
		return c.JSON(http.StatusBadRequest, exception.BadRequest("image type is not valid. please upload *.jpg or *jpeg"))
	}

	// load aws config from terminal
	cfg, err := config.LoadDefaultConfig(c.Request().Context(), config.WithRegion(AWS_REGION))
	if err != nil {
		log.Println("cannot load aws config, because: ", err.Error())
		return c.JSON(http.StatusInternalServerError, exception.ServerError("server sibuk"))
	}

	//load s3 config
	client := s3.NewFromConfig(cfg)

	_, err = client.PutObject(c.Request().Context(), &s3.PutObjectInput{
		Bucket: aws.String(AWS_S3_BUCKET_NAME),
		Key:    aws.String(key),
		Body:   src,
		ACL:    types.ObjectCannedACLPublicRead,
	})

	if err != nil {
		log.Printf("Couldn't upload file. Here's why: %v\n\n", err)
		return c.JSON(http.StatusInternalServerError, exception.ServerError("Server sibuk"))
	}

	return c.JSON(200, &converter.ImageResponse{
		Message: "File uploaded sucessfully",
		Data: &converter.ImageUrl{
			ImageUrl: "https://" + AWS_S3_BUCKET_NAME + ".s3.amazonaws.com/" + key,
		},
	})
}

func (i *ImageHandler) isValidImage(file multipart.File) (bool, error) {
	buffer := make([]byte, 512)
	if _, err := file.Read(buffer); err != nil {
		return false, err
	}
	// Reset the read pointer of the file
	if _, err := file.Seek(0, 0); err != nil {
		return false, err
	}
	filetype := http.DetectContentType(buffer)
	return filetype == "image/jpeg" || filetype == "image/jpg", nil
}
