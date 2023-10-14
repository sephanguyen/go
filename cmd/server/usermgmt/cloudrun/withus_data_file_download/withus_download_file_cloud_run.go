package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"cloud.google.com/go/storage"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/crypto/ssh"
)

const (
	withusKeyFile = "/secret-manager/withus-key-management"
)

// Config represents parameters from CloudRun
// https://cloud.google.com/run/docs/quickstarts/jobs/build-create-go
type Config struct {
	// Job-defined
	taskNum    string
	attemptNum string

	// User-defined
	ipAddress             string
	port                  string
	username              string
	password              string
	filePath              string
	fileName              string
	fileNamePrefix        string
	fileUploadDate        time.Time
	testFilePath          string
	testFileName          string
	gcloudProjectID       string
	gcloudStorageBucket   string
	gcloudStorageFilePath string
	privateKeyBase64      string
}

func configFromEnv() (Config, error) {
	// Job-defined
	taskNum := os.Getenv("CLOUD_RUN_TASK_INDEX")
	attemptNum := os.Getenv("CLOUD_RUN_TASK_ATTEMPT")

	// User-defined
	ipAddress := os.Getenv("SERVER_IP")
	port := os.Getenv("SERVER_PORT")
	username := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")
	privateKeyBase64 := os.Getenv("PRIVATE_KEY_BASE64")
	filePath := os.Getenv("FILE_PATH")
	fileName := os.Getenv("FILE_NAME")
	fileNamePrefix := os.Getenv("FILE_NAME_PREFIX")
	testFilePath := os.Getenv("TEST_FILE_PATH")
	testFileName := os.Getenv("TEST_FILE_NAME")
	gcloudProjectID := os.Getenv("GCLOUD_PROJECT_ID")
	gcloudStorageBucket := os.Getenv("GCLOUD_STORAGE_BUCKET")
	gcloudStorageFilePath := os.Getenv("GCLOUD_STORAGE_FILE_PATH")

	tokyoLocation, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return Config{}, errors.Wrap(err, "time.LoadLocation()")
	}

	fileUploadDate := time.Now().In(tokyoLocation)
	if fileUploadDateString := os.Getenv("FILE_UPLOAD_DATE"); fileUploadDateString != "" {
		// FILE_UPLOAD_DATE will be input as JSP timezone
		parsedTime, err := time.ParseInLocation(TimeFormatYYYYMMDD, fileUploadDateString, tokyoLocation)
		if err != nil {
			return Config{}, errors.Wrap(err, "failed to parse FILE_UPLOAD_DATE")
		}
		fileUploadDate = parsedTime
	}

	config := Config{
		taskNum:               taskNum,
		attemptNum:            attemptNum,
		ipAddress:             ipAddress,
		port:                  port,
		username:              username,
		password:              password,
		filePath:              filePath,
		fileName:              fileName,
		fileNamePrefix:        fileNamePrefix,
		fileUploadDate:        fileUploadDate,
		testFilePath:          testFilePath,
		testFileName:          testFileName,
		gcloudProjectID:       gcloudProjectID,
		gcloudStorageBucket:   gcloudStorageBucket,
		gcloudStorageFilePath: gcloudStorageFilePath,
		privateKeyBase64:      privateKeyBase64,
	}

	return config, nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	zapLogger := NewZapLogger("debug", false)
	ctx = ctxzap.ToContext(ctx, zapLogger)

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	err := godotenv.Load(withusKeyFile)
	if err != nil {
		zapLogger.Fatal("failed to load key", zap.Error(err))
		return
	}

	config, err := configFromEnv()
	if err != nil {
		zapLogger.Fatal("failed connect client to get config", zap.Error(err))
		return
	}

	zapLogger.Info("config",
		zap.String("ip", config.ipAddress),
		zap.String("port", config.port),
		zap.String("username", config.username),
		// zap.String("password", string(config.password[0])+"*****"+string(config.password[len(config.password)-1])),
		zap.String("filePath", config.filePath),
		zap.String("fileName", config.fileName),
		zap.String("fileNamePrefix", config.fileNamePrefix),
		zap.String("fileUploadDate", config.fileUploadDate.Format(TimeFormatYYYYMMDD)),
		zap.String("testFilePath", config.testFilePath),
		zap.String("gcloudProjectID", config.gcloudProjectID),
		zap.String("gcloudStorageBucket", config.gcloudStorageBucket),
		zap.String("gcloudStorageFilePath", config.gcloudStorageFilePath),
	)

	/*client, err := storage.NewClient(ctx)
	if err != nil {
		zapLogger.Fatal("failed connect client to gcloud storage", zap.Error(err))
	}
	defer client.Close()

	bucket := client.Bucket(config.gcloudBucket)
	testObj := bucket.Object("withus/test.txt")
	testObjWriter := testObj.NewWriter(ctx)

	if _, err := testObjWriter.Write([]byte(`test`)); err != nil {
		zapLogger.Fatal("failed to write data to obj", zap.Error(err))
	}
	if err := testObjWriter.Close(); err != nil {
		zapLogger.Fatal("failed to close testObjWriter", zap.Error(err))
	}
	zapLogger.Info("wrote")*/

	authMethods := make([]ssh.AuthMethod, 0)

	if privateKey := config.privateKeyBase64; strings.TrimSpace(privateKey) != "" {
		privateKeyText, err := base64.StdEncoding.DecodeString(config.privateKeyBase64)
		if err != nil {
			zapLogger.Panic("failed to decode base64", zap.Error(err))
		}
		signerKey, err := ssh.ParsePrivateKey(privateKeyText)
		if err != nil {
			zapLogger.Panic("failed to parse signer key", zap.Error(err))
		}
		authMethods = append(authMethods, ssh.PublicKeys(signerKey))
	}

	if password := config.password; strings.TrimSpace(password) != "" {
		authMethods = append(authMethods, ssh.Password(config.password))
		authMethods = append(authMethods, ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) ([]string, error) {
			// Just send the password back for all questions
			answers := make([]string, len(questions))
			for i, _ := range answers {
				answers[i] = password // replace this
			}

			return answers, nil
		}))
	}

	clientConfig := &ssh.ClientConfig{
		User:            config.username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         15 * time.Second,
	}

	sshClient, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", config.ipAddress, config.port), clientConfig)
	if err != nil {
		zapLogger.Fatal(
			"failed to dial to server",
			zap.Error(err),
		)
	}
	defer sshClient.Close()

	zapLogger.Info(fmt.Sprintf("connected to %s", config.ipAddress))

	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		zapLogger.Fatal(
			"failed to create new sftp client",
			zap.Error(err),
		)
	}
	defer sftpClient.Close()

	// Download files from withus then upload to gcloud storage
	client, err := storage.NewClient(ctx)
	if err != nil {
		zapLogger.Fatal("failed connect client to gcloud storage", zap.Error(err))
	}
	defer client.Close()

	bucket := client.Bucket(config.gcloudStorageBucket).UserProject(config.gcloudProjectID)
	fileName := FileNameToDownload(config.fileNamePrefix, config.fileUploadDate)

	if err := downloadThenUpload(ctx, sftpClient, config.filePath, fileName, bucket, config.gcloudStorageFilePath); err != nil {
		zapLogger.Fatal(
			"failed to download/upload file",
			zap.Error(err),
			zap.String("filePath", config.filePath),
			zap.String("fileName", fileName),
		)
	}

	zapLogger.Info("downloaded and uploaded file successfully")
}

func downloadThenUpload(ctx context.Context, srcSFTPClient *sftp.Client, srcFilePath string, srcFileName string, destBucket *storage.BucketHandle, destFilePath string) error {
	zapLogger := ctxzap.Extract(ctx)

	/*files, err := srcSFTPClient.ReadDir(srcFilePath)
	if err != nil {
		return errors.Wrap(err, "failed to read dir")
	}

	text := strings.Builder{}
	text.WriteString("fileName,fileSize,fileModTime")

	for _, file := range files {
		text.WriteString(`/n`)
		text.WriteString(file.Name())
		text.WriteString(`,`)
		text.WriteString(file.ModTime().UTC().String())
		text.WriteString(`,`)
		text.WriteString(fmt.Sprintf("%v bytes", file.Size()))

		zapLogger.Info(
			file.Name(),
			zap.Time("modTime", file.ModTime()),
			zap.Int64("size", file.Size()),
		)
	}

	zapLogger.Info(text.String())*/

	file, err := srcSFTPClient.Open(srcFilePath + "/" + srcFileName)
	if err != nil {
		return errors.Wrap(err, "failed to open file")
	}
	defer file.Close()

	obj := destBucket.Object(destFilePath + "/" + srcFileName)
	objWriter := obj.NewWriter(ctx)
	defer func() {
		_ = objWriter.Close()
		zapLogger.Info(
			"uploaded file",
			zap.String("srcFilePath", srcFilePath),
			zap.String("srcFileName", srcFileName),
			zap.String("destFilePath", destFilePath),
		)
	}()

	if _, err = io.Copy(objWriter, file); err != nil {
		return errors.Wrap(err, "failed to write data to obj")
	}
	if err := objWriter.Close(); err != nil {
		return errors.Wrap(err, "failed to close obj write")
	}

	return nil
}

/*
	func downloadFile(sftpClient *sftp.Client, filePath string, fileName string) ([]byte, error) {
		file, err := sftpClient.Open(filePath + fileName)
		if err != nil {
			return nil, errors.Wrap(err, "failed to open file")
		}
		defer file.Close()

		stat, err := file.Stat()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get file stat")
		}

		data := make([]byte, 0, stat.Size())
		if _, err := file.Read(data); err != nil {
			return nil, errors.Wrap(err, "failed to read file")
		}

		return data, nil
	}

	func uploadFileToGcloudStorage(ctx context.Context, bucket *storage.BucketHandle, fileName string, data []byte) error {
		obj := bucket.Object(fileName)
		objWriter := obj.NewWriter(ctx)
		defer objWriter.Close()

		if _, err := objWriter.Write(data); err != nil {
			return errors.Wrap(err, "failed to write data to obj")
		}
		if err := objWriter.Close(); err != nil {
			return errors.Wrap(err, "failed to close obj write")
		}

		return nil
	}
*/

const TimeFormatYYYYMMDD = "20060102"

const (
	FileNameInfix     = "_users"
	FileNameExtension = ".tsv"
)

func FileNameToDownload(fileNamePrefix string, fileUploadTime time.Time) string {
	return fmt.Sprintf("%s%s%s%s", fileNamePrefix, FileNameInfix, FileNameSuffix(fileUploadTime), FileNameExtension)
}

func FileNameSuffix(fileUploadTime time.Time) string {
	return fileUploadTime.Format(TimeFormatYYYYMMDD)
}

var (
	stdout zapcore.WriteSyncer = os.Stdout
	stderr zapcore.WriteSyncer = os.Stderr
	atom                       = zap.NewAtomicLevelAt(zap.ErrorLevel)
)

func NewZapLogger(logLevel string, isLocalEnv bool) *zap.Logger {
	var (
		zapLogger *zap.Logger
		zapLogLvl zapcore.Level
	)
	err := zapLogLvl.Set(logLevel)
	if err != nil {
		log.Println("(warning) failed to parse logLevel:", err.Error())
		zapLogLvl = zap.WarnLevel
	}
	atom.SetLevel(zapLogLvl)
	if isLocalEnv {
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.Level = atom
		zapLogger, err := config.Build()
		if err != nil {
			panic(fmt.Errorf("failed to build zap logger: %s", err))
		}
		return zapLogger
	}

	// Configure console output.
	consoleEncoder := newJSONEncoder()

	// We use zapcore.NewTee to direct logs of different levels to different outputs
	// while also respecting the log level from atom.
	core := zapcore.NewTee(
		zapcore.NewCore(
			consoleEncoder,
			zapcore.Lock(stderr),
			zap.LevelEnablerFunc(func(l zapcore.Level) bool {
				return l >= zapcore.ErrorLevel && atom.Enabled(l)
			}),
		),
		zapcore.NewCore(
			consoleEncoder,
			zapcore.Lock(stdout),
			zap.LevelEnablerFunc(func(l zapcore.Level) bool {
				return l < zapcore.ErrorLevel && atom.Enabled(l)
			}),
		),
	)
	zapLogger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	zap.RedirectStdLog(zapLogger)
	return zapLogger
}

// Create a new JSON log encoder with the correct settings.
func newJSONEncoder() zapcore.Encoder {
	return zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "severity",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		EncodeLevel:    appendLogLevel,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})
}

func appendLogLevel(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	switch l {
	case zapcore.DebugLevel:
		enc.AppendString("debug")
	case zapcore.InfoLevel:
		enc.AppendString("info")
	case zapcore.WarnLevel:
		enc.AppendString("warning")
	case zapcore.ErrorLevel:
		enc.AppendString("error")
	case zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		enc.AppendString("critical")
	default:
		enc.AppendString(fmt.Sprintf("Level(%d)", l))
	}
}
