package aws

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestClientStatusUpdater_NewFile(t *testing.T) {
	stubFileObj := "name:{img.png},dataUrl:{data:image/png;base64,iVBggg==}"
	stubFileObjBad := "dataUrl:{data:image/png;base64,iVBORw0KGgoA=="
	// arrange
	cases := []struct {
		desc        string
		stubFileObj *string
		wantFile    File
		wantErr     error
	}{
		{
			desc:        "Should returns error when fileParams != 3",
			stubFileObj: &stubFileObjBad,
			wantFile:    File{},
			wantErr:     errors.New("invalid file object"),
		},
		{
			desc:        "Should returns no error and file",
			stubFileObj: &stubFileObj,
			wantFile: File{
				content:  "data:image/png;base64,iVBggg==",
				fileName: "img.png",
			},
			wantErr: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			// actual
			_, gotErr := NewFile(c.stubFileObj)

			// assert
			assert.Equal(t, c.wantErr, gotErr)
		})
	}
}

func TestClientStatusUpdater_NewAWSConnector(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svcStub := NewMockiS3Client(ctrl)
	generatorStub := NewMockiGenerate(ctrl)

	// arrange
	cases := []struct {
		desc             string
		awsInfo          AWSInfo
		timeout          time.Duration
		svc              *MockiS3Client
		generator        *MockiGenerate
		wantAWSConnector *AWSConnector
		wantErr          error
	}{
		{
			desc:             "Should returns error when bucket is empty",
			awsInfo:          AWSInfo{},
			svc:              NewMockiS3Client(ctrl),
			generator:        NewMockiGenerate(ctrl),
			wantAWSConnector: nil,
			wantErr:          errors.New("bucket is empty"),
		},
		{
			desc: "Should returns error when aws url is empty",
			awsInfo: AWSInfo{
				Bucket: "test",
			},
			svc:              NewMockiS3Client(ctrl),
			generator:        NewMockiGenerate(ctrl),
			wantAWSConnector: nil,
			wantErr:          errors.New("aws url is empty"),
		},
		{
			desc: "Should returns error when timeout is negative",
			awsInfo: AWSInfo{
				Bucket: "test",
				URL:    "test.com",
			},
			svc:              NewMockiS3Client(ctrl),
			generator:        NewMockiGenerate(ctrl),
			wantAWSConnector: nil,
			wantErr:          errors.New("timeout is negative"),
		},
		{
			desc: "Should returns no error and returns *AWSConnector",
			awsInfo: AWSInfo{
				Bucket: "test",
				URL:    "test.com",
			},
			timeout:   3 * time.Minute,
			svc:       svcStub,
			generator: generatorStub,
			wantAWSConnector: &AWSConnector{
				timeout: time.Minute * 3,
				AWSInfo: AWSInfo{
					Bucket: "test",
					URL:    "test.com",
				},
				svc:       svcStub,
				generator: generatorStub,
			},
			wantErr: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			// actual
			got, gotErr := NewAWSConnector(c.awsInfo, c.timeout, c.svc, c.generator)

			// assert
			assert.Equal(t, c.wantAWSConnector, got)
			assert.Equal(t, c.wantErr, gotErr)
		})
	}
}

func TestClientStatusUpdater_PutFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	stubFileObj := "name:{img.png},dataUrl:{data:image/png;base64,iVBggg==}"
	stubFileObjBad := "dataUrl:{data:image/png;base64,iVBORw0KGgoA=="
	stubBadContent := "name:{img.png},dataUrl:{data:image/png;}"

	// arrange
	cases := []struct {
		desc               string
		fileObj            *string
		svc                *MockiS3Client
		generator          *MockiGenerate
		wantUniqueFileName string
		wantErr            error
	}{
		{
			desc: "Should returns error when NewFile failed",

			fileObj:            &stubFileObjBad,
			svc:                NewMockiS3Client(ctrl),
			generator:          NewMockiGenerate(ctrl),
			wantUniqueFileName: "",
			wantErr:            errors.New("invalid file object"),
		},
		{
			desc:               "Should returns error when dataurl.DecodeString failed",
			fileObj:            &stubBadContent,
			svc:                NewMockiS3Client(ctrl),
			generator:          NewMockiGenerate(ctrl),
			wantUniqueFileName: "",
			wantErr:            errors.New("unterminated parameter sequence"),
		},
		{
			desc:    "Should returns error when PutObjectWithContext filed",
			fileObj: &stubFileObj,
			generator: func(m *MockiGenerate) *MockiGenerate {
				m.EXPECT().GenerateTime().Return("time")
				m.EXPECT().GenerateUUID().Return("111")
				return m
			}(NewMockiGenerate(ctrl)),
			svc: func(m *MockiS3Client) *MockiS3Client {
				m.EXPECT().PutObjectWithContext(gomock.Any(), gomock.Any()).
					Return(errors.New("AWS returned error, saving file failed"))
				return m
			}(NewMockiS3Client(ctrl)),
			wantUniqueFileName: "",
			wantErr:            errors.New("AWS returned error, saving file failed"),
		},
		{
			desc:    "Should returns no error",
			fileObj: &stubFileObj,
			generator: func(m *MockiGenerate) *MockiGenerate {
				m.EXPECT().GenerateTime().Return("time")
				m.EXPECT().GenerateUUID().Return("111")
				return m
			}(NewMockiGenerate(ctrl)),
			svc: func(m *MockiS3Client) *MockiS3Client {
				m.EXPECT().PutObjectWithContext(gomock.Any(), gomock.Any()).
					Return(nil)
				return m
			}(NewMockiS3Client(ctrl)),
			wantUniqueFileName: "time_111_img.png",
			wantErr:            nil,
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			// actual
			awsInfo := AWSInfo{
				Bucket: "test bucket",
				URL:    "test URL",
				Region: "test region",
				ID:     "123",
				Secret: "secret",
				Token:  "",
			}
			aws, _ := NewAWSConnector(awsInfo, time.Minute, c.svc, c.generator)
			got, gotErr := aws.PutFile(c.fileObj)

			// assert
			assert.Equal(t, c.wantUniqueFileName, got)
			assert.Equal(t, c.wantErr, gotErr)
		})
	}
}

func TestClientStatusUpdater_SetBucketReadOnlyPolicy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// arrange
	cases := []struct {
		desc      string
		svc       *MockiS3Client
		generator *MockiGenerate
		wantErr   error
	}{
		{
			desc: "Should returns error when PutBucketPolicy failed",
			svc: func(m *MockiS3Client) *MockiS3Client {
				m.EXPECT().PutBucketPolicy(gomock.Any()).Return(errors.New("test error"))
				return m
			}(NewMockiS3Client(ctrl)),
			generator: NewMockiGenerate(ctrl),
			wantErr:   errors.New("test error"),
		},
		{
			desc: "Should returns no error",
			svc: func(m *MockiS3Client) *MockiS3Client {
				m.EXPECT().PutBucketPolicy(gomock.Any()).Return(nil)
				return m
			}(NewMockiS3Client(ctrl)),
			generator: NewMockiGenerate(ctrl),
			wantErr:   nil,
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			// actual
			awsInfo := AWSInfo{
				Bucket: "test bucket",
				URL:    "test URL",
				Region: "test region",
				ID:     "123",
				Secret: "secret",
				Token:  "",
			}
			aws, _ := NewAWSConnector(awsInfo, time.Minute, c.svc, c.generator)
			gotErr := aws.SetBucketReadOnlyPolicy()

			// assert
			assert.Equal(t, c.wantErr, gotErr)
		})
	}
}
