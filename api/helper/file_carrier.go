// InfoMark - a platform for managing courses with
//            distributing exercise sheets and testing exercise submissions
// Copyright (C) 2019  ComputerGraphics Tuebingen
// Authors: Patrick Wieschollek
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package helper

import (
  "crypto/sha256"
  "errors"
  "fmt"
  "io"
  "mime/multipart"
  "net/http"
  "os"
  pathpkg "path"
  "strconv"
  "strings"

  "github.com/spf13/viper"
)

// FaileCarrier is a unified way to handle uploads and downloads of different
// files.

type FileCategory int32

const (
  AvatarCategory                FileCategory = 0
  SheetCategory                 FileCategory = 1
  PublicTestCategory            FileCategory = 2
  PrivateTestCategory           FileCategory = 3
  MaterialCategory              FileCategory = 4
  SubmissionCategory            FileCategory = 5
  SubmissionsCollectionCategory FileCategory = 6
)

type FileManager interface {
  WriteToBody(w http.ResponseWriter) error
  WriteToDisk(req multipart.File) error
  GetContentType() (string, error)
  Path(fallback bool) bool
  Delete() error
  Exists() bool
}

// FileHandle represents all information for file being uploaded or downloaded.
type FileHandle struct {
  Category   FileCategory
  ID         int64    // an unique identifier (e.g. from database)
  Extensions []string //
  MaxBytes   int64    // 0 means no limit
  Infos      []int64
}

// NewAvatarFileHandle will handle user avatars. We support jpg only.
func NewAvatarFileHandle(userID int64) *FileHandle {
  return &FileHandle{
    Category:   AvatarCategory,
    ID:         userID,
    Extensions: []string{"jpg", "jpeg", "png"},
    MaxBytes:   viper.GetInt64("max_request_avatar_bytes"),
  }
}

// NewSheetFileHandle will handle exercise sheets (zip files).
func NewSheetFileHandle(ID int64) *FileHandle {
  return &FileHandle{
    Category:   SheetCategory,
    ID:         ID,
    Extensions: []string{"zip"},
    MaxBytes:   0,
  }
}

// NewPublicTestFileHandle will handle the testing framework for
// public unit tests (zip files).
func NewPublicTestFileHandle(ID int64) *FileHandle {
  return &FileHandle{
    Category:   PublicTestCategory,
    ID:         ID,
    Extensions: []string{"zip"},
    MaxBytes:   0,
  }
}

// NewPrivateTestFileHandle will handle the testing framework for
// private unit tests (zip files).
func NewPrivateTestFileHandle(ID int64) *FileHandle {
  return &FileHandle{
    Category:   PrivateTestCategory,
    ID:         ID,
    Extensions: []string{"zip"},
    MaxBytes:   0,
  }
}

// NewMaterialFileHandle will handle course slides or extra material (zip files).
func NewMaterialFileHandle(ID int64) *FileHandle {
  return &FileHandle{
    Category:   MaterialCategory,
    ID:         ID,
    Extensions: []string{"zip", "pdf"},
    MaxBytes:   0,
  }
}

// NewSubmissionFileHandle will handle homework/exercise submissiosn (zip files).
func NewSubmissionFileHandle(ID int64) *FileHandle {
  return &FileHandle{
    Category:   SubmissionCategory,
    ID:         ID,
    Extensions: []string{"zip"},
    MaxBytes:   viper.GetInt64("max_request_submission_bytes"),
  }
}

// NewSubmissionFileHandle will handle homework/exercise submissiosn (zip files).
func NewSubmissionsCollectionFileHandle(courseID int64, sheetID int64, taskID int64, groupID int64) *FileHandle {
  return &FileHandle{
    Category:   SubmissionsCollectionCategory,
    ID:         0,
    Extensions: []string{"zip"},
    MaxBytes:   0,
    Infos:      []int64{courseID, sheetID, taskID, groupID},
  }
}

// Path returns a path without checking if it exists.
func (f *FileHandle) Sha256() (string, error) {

  hnd, err := os.Open(f.Path())
  if err != nil {
    return "", err
  }
  defer hnd.Close()

  h := sha256.New()
  if _, err := io.Copy(h, hnd); err != nil {
    return "", err
  }

  return fmt.Sprintf("%x", h.Sum(nil)), nil

}

func (f *FileHandle) Path() string {
  switch f.Category {
  case AvatarCategory:

    for _, ext := range f.Extensions {
      path := fmt.Sprintf("%s/avatars/%d.%s", viper.GetString("uploads_dir"), f.ID, ext)
      if FileExists(path) {
        return path
      }
    }
    return ""

  case SheetCategory:
    return fmt.Sprintf("%s/sheets/%d.zip", viper.GetString("uploads_dir"), f.ID)

  case PublicTestCategory:
    return fmt.Sprintf("%s/tasks/%d-public.zip", viper.GetString("uploads_dir"), f.ID)

  case PrivateTestCategory:
    return fmt.Sprintf("%s/tasks/%d-private.zip", viper.GetString("uploads_dir"), f.ID)

  case MaterialCategory:

    for _, ext := range f.Extensions {
      path := fmt.Sprintf("%s/materials/%d.%s", viper.GetString("uploads_dir"), f.ID, ext)
      if FileExists(path) {
        return path
      }
    }
    return ""

  case SubmissionCategory:
    return fmt.Sprintf("%s/submissions/%d.zip", viper.GetString("uploads_dir"), f.ID)
  case SubmissionsCollectionCategory:
    return fmt.Sprintf("%s/collection-course%d-sheet%d-task%d-group%d.zip",
      viper.GetString("generated_files_dir"), f.Infos[0], f.Infos[1], f.Infos[2], f.Infos[3])
  }
  return ""
}

// Exists checks if a file really exists.
func FileExists(path string) bool {
  if _, err := os.Stat(path); os.IsNotExist(err) {
    return false
  }

  return true
}

// FileTouch creates an empty file
func FileTouch(path string) error {
  emptyFile, err := os.Create(path)
  defer emptyFile.Close()
  return err
}

// FileDelete deletes an file
func FileDelete(path string) error {
  return os.Remove(path)
}

// Exists checks if a file really exists.
func (f *FileHandle) Exists() bool {
  return FileExists(f.Path())
}

// Delete deletes a file from disk.
func (f *FileHandle) Delete() error {
  return os.Remove(f.Path())
}

// GetContentType tries to predict the content type without reading the entire
// file. There are some issues with this function as it cannot distinguish
// between zip and octstream.
func (f *FileHandle) GetContentType() (string, error) {

  // Only the first 512 bytes are used to sniff the content type.
  buffer := make([]byte, 512)

  file, err := os.Open(f.Path())
  if err != nil {
    return "", err
  }
  defer file.Close()

  _, err = file.Read(buffer)
  if err != nil {
    return "", err
  }

  // Use the net/http package's handy DectectContentType function. Always returns a valid
  // content-type by returning "application/octet-stream" if no others seemed to match.
  contentType := http.DetectContentType(buffer)

  return contentType, nil
}

type DummyWriter struct{}

func (h DummyWriter) Header() http.Header {
  return make(map[string][]string)
}

func (h DummyWriter) Write([]byte) (int, error) {
  return 0, nil
}

func (h DummyWriter) WriteHeader(statusCode int) {}

// WriteToBody will write a file from disk to the http reponse (download process)
func (f *FileHandle) WriteToBody(w http.ResponseWriter) error {

  // check if file exists
  file, err := os.Open(f.Path())
  if err != nil {
    return err
  }
  defer file.Close()

  path_split := strings.Split(f.Path(), "/")
  publicFilename := fmt.Sprintf("%s-%s", path_split[len(path_split)-2], path_split[len(path_split)-1])

  w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=infomark-%s", publicFilename))

  // prepare header
  fileType, err := f.GetContentType()
  if err != nil {
    return err
  }
  w.Header().Set("Content-Type", fileType)

  // return file
  _, err = io.Copy(w, file)
  if err != nil {
    return err
  }

  return nil
}

func (f *FileHandle) WriteToBodyWithName(publicFilename string, w http.ResponseWriter) error {

  // check if file exists
  file, err := os.Open(f.Path())
  if err != nil {
    return err
  }
  defer file.Close()

  w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", publicFilename))

  // prepare header
  fileType, err := f.GetContentType()
  if err != nil {
    return err
  }
  w.Header().Set("Content-Type", fileType)

  // return file
  _, err = io.Copy(w, file)
  if err != nil {
    return err
  }

  return nil
}

// Check if file is zip file based on magic number
func IsZipFile(buf []byte) bool {
  return len(buf) > 3 &&
    buf[0] == 0x50 && buf[1] == 0x4B &&
    (buf[2] == 0x3 || buf[2] == 0x5 || buf[2] == 0x7) &&
    (buf[3] == 0x4 || buf[3] == 0x6 || buf[3] == 0x8)
}

// Check if file is pdf file based on magic number
func IsPdfFile(buf []byte) bool {
  return len(buf) > 3 &&
    buf[0] == 0x25 && buf[1] == 0x50 &&
    buf[2] == 0x44 && buf[3] == 0x46
}

// Check if file is jpg file based on magic number
func IsJpegFile(buf []byte) bool {
  return len(buf) > 2 &&
    buf[0] == 0xFF &&
    buf[1] == 0xD8 &&
    buf[2] == 0xFF
}

// Check if file is png file based on magic number
func IsPngFile(buf []byte) bool {
  return len(buf) > 3 &&
    buf[0] == 0x89 && buf[1] == 0x50 &&
    buf[2] == 0x4E && buf[3] == 0x47
}

// WriteToDisk will save uploads from a http request to the directory specified
// in the config.
func (f *FileHandle) WriteToDisk(r *http.Request, fieldName string) (string, error) {

  w := DummyWriter{}

  if f.MaxBytes != 0 {
    r.Body = http.MaxBytesReader(w, r.Body, f.MaxBytes)
  }

  // receive data from post request
  if err := r.ParseMultipartForm(32 << 20); err != nil {
    return "", err
  }

  // we are interested in the field "file_data"
  file, handler, err := r.FormFile(fieldName)
  if err != nil {
    return "", err
  }
  defer file.Close()

  path := f.Path()

  // Extract magic number from file
  file_magic := make([]byte, 4)
  if n, err := file.Read(file_magic); err != nil || n != 4 {
    return "", errors.New("Unable to extract 4 Bytes for magic number determination.")
  }
  if n, err := file.Seek(0, io.SeekStart); n != 0 || err != nil {
    return "", errors.New("Fail to seek to beginning of file.")
  }

  switch f.Category {
  case AvatarCategory:
    path_to_delete := fmt.Sprintf("%s/avatars/%s.png", viper.GetString("uploads_dir"), strconv.FormatInt(f.ID, 10))
    FileDelete(path_to_delete)
    path_to_delete = fmt.Sprintf("%s/avatars/%s.jpg", viper.GetString("uploads_dir"), strconv.FormatInt(f.ID, 10))
    FileDelete(path_to_delete)
    if IsJpegFile(file_magic) {
      path = fmt.Sprintf("%s/avatars/%s.jpg", viper.GetString("uploads_dir"), strconv.FormatInt(f.ID, 10))
    } else if IsPngFile(file_magic) {
      path = fmt.Sprintf("%s/avatars/%s.png", viper.GetString("uploads_dir"), strconv.FormatInt(f.ID, 10))
    } else {
      return "", errors.New("We support JPG/JPEG/PNG files only.")
    }

  case SheetCategory,
    PublicTestCategory,
    PrivateTestCategory,
    SubmissionCategory:
    if !IsZipFile(file_magic) {
      return "", errors.New("We support ZIP files only. But the given file is no Zip file!")
    }
  case MaterialCategory:
    // delete both possible files
    // ids are unique. Hence we only delete the file associated with the id
    path_to_delete := fmt.Sprintf("%s/materials/%s.zip", viper.GetString("uploads_dir"), strconv.FormatInt(f.ID, 10))
    FileDelete(path_to_delete)
    path_to_delete = fmt.Sprintf("%s/materials/%s.pdf", viper.GetString("uploads_dir"), strconv.FormatInt(f.ID, 10))
    FileDelete(path_to_delete)

    if IsPdfFile(file_magic) {
      path = fmt.Sprintf("%s/materials/%s.pdf", viper.GetString("uploads_dir"), strconv.FormatInt(f.ID, 10))
    } else if IsZipFile(file_magic) {
      path = fmt.Sprintf("%s/materials/%s.zip", viper.GetString("uploads_dir"), strconv.FormatInt(f.ID, 10))
    } else {
      return "", errors.New("Only PDF and ZIP files are allowed.")
    }
  }

  // delete path
  FileDelete(path)
  // try to open new file
  hnd, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)
  if err != nil {
    return "", err
  }
  defer hnd.Close()

  // copy file from request
  _, err = io.Copy(hnd, file)
  return pathpkg.Base(handler.Filename), err

}
