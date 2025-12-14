package imgutils

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"multipass/pkg/apperror"
	"multipass/pkg/common"
	"multipass/pkg/logging"
	"multipass/pkg/utils"
)

// isValidFileExtension checks if the file has a valid extension.
func IsValidFileExtension(ext string) bool {
	return strings.Contains(common.AllowedExtensions, strings.TrimPrefix(ext, "."))
}

func IsValidImgMIMEType(file multipart.File, log logging.Logger) (bool, error) {
	buf := make([]byte, 512)
	_, err := file.Read(buf)
	if err != nil {
		return false, apperror.ErrUnsupportedFileType(err, log, nil)
	}

	// DETECT MIME TYPE FROM FIRST 512 BYTES
	mimeType := http.DetectContentType(buf)
	validMimeTypes := []string{"image/jpeg", "image/png", "image/jgp", "image/webp"}

	// CHECK MIME TYPE WITH THE LIST OF ALLOWED TYPES
	for _, validType := range validMimeTypes {
		if mimeType == validType {
			return true, nil
		}
	}

	return false, fmt.Errorf("invalid MIME type: %s", mimeType)
}

func UploadProfilePicture(file *multipart.FileHeader, userID int, log logging.Logger) (string, error) {
	fileExtension := strings.ToLower(filepath.Ext(file.Filename))
	metaData := common.Envelop{
		"file_size": file.Size,
		"file_ext":  fileExtension,
		"user_id":   userID,
	}

	// STEP 1: VALIDATE FILE SIZE
	if file.Size > int64(common.MaxFileSize) {
		return "", apperror.ErrFileTooLarge(errors.New(apperror.ErrFileTooLargeMsg), log, metaData)
	}

	// STEP 2: VALIDATE FILE FORMAT
	if !IsValidFileExtension(fileExtension) {
		return "", apperror.ErrUnsupportedFileType(errors.New(apperror.ErrUnsupportedFileTypeMsg), log, metaData)
	}

	profileImgPath := os.Getenv("USER_PROFILE_IMG_PATH")
	// STEP 3: ENSURE DIRECTORY PATH EXISTS
	err := utils.EnsureDirExists(profileImgPath)
	if err != nil {
		return "", apperror.ErrDirectoryNotFound(err, log, metaData)
	}

	// STEP 4: GENERATE UNIQUE FILENAME & FILEPATH
	uniqueFilename := fmt.Sprintf("%d_%s%s", userID, time.Now().Format("20060102150405"), fileExtension)
	filePath := filepath.Join(profileImgPath, uniqueFilename)

	metaData["file_path"] = filePath

	// STEP 5: OPEN FILE FROM MULTIPART HEADER
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("could not open file: %w", err)
	}

	defer src.Close()

	// STEP 6: CREATE A DESTINATION FILE
	dest, err := os.Create(filePath)
	if err != nil {
		return "", apperror.ErrFileCreateFailed(err, log, metaData)
	}

	defer dest.Close()

	// STEP 7: COPY FILE CONTENT TO DESTINATION WITH BUFFERING
	_, err = io.Copy(dest, src)
	if err != nil {
		return "", apperror.ErrFileCopyFailed(err, log, metaData)
	}

	// STEP 8: CONSTRUCT PUBLIC URL FOR FILE
	var s strings.Builder
	s.WriteString(os.Getenv("USER_PROFILE_IMG_BASE"))
	s.WriteString(uniqueFilename)
	profileImgUrl := s.String()

	return profileImgUrl, nil
}
