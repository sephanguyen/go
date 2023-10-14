package agoratokenbuilder

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"hash/crc32"
	"io"
	"math/rand"
	"sort"
	"time"
)

const VERSION_LENGTH = 3
const APP_ID_LENGTH = 32

type Privileges uint16

const (
	KJoinChannel        = 1
	KPublishAudioStream = 2
	KPublishVideoStream = 3
	KPublishDataStream  = 4

	KPublishAudiocdn           = 5
	KPublishVideoCdn           = 6
	KRequestPublishAudioStream = 7
	KRequestPublishVideoStream = 8
	KRequestPublishDataStream  = 9
	KInvitePublishAudioStream  = 10
	KInvitePublishVideoStream  = 11
	KInvitePublishDataStream   = 12

	KAdministrateChannel = 101
	KLoginRtm            = 1000
)

type AccessToken struct {
	AppID          string
	AppCertificate string
	ChannelName    string
	UidStr         string
	Ts             uint32
	Salt           uint32
	Message        map[uint16]uint32
	Signature      string
	CrcChannelName uint32
	CrcUid         uint32
	MsgRawContent  string
}

func random(min int, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min) + min
}

func getVersion() string {
	return "006"
}

// CreateAccessToken2 copied from Agora example
func CreateAccessToken2(appID, appCertificate, channelName string, uid string) AccessToken {
	ts := uint32(time.Now().Unix()) + 24*3600
	salt := uint32(random(1, 99999999))
	message := make(map[uint16]uint32)
	return AccessToken{appID, appCertificate, channelName, uid, ts, salt, message, "", 0, 0, ""}
}

func (token *AccessToken) AddPrivilege(privilege Privileges, expireTimestamp uint32) {
	pri := uint16(privilege)
	token.Message[pri] = expireTimestamp
}

func (token *AccessToken) Build() (string, error) {
	ret := ""
	version := getVersion()

	buf_m := new(bytes.Buffer)
	if err := packUint32(buf_m, token.Salt); err != nil {
		return ret, err
	}
	if err := packUint32(buf_m, token.Ts); err != nil {
		return ret, err
	}
	if err := packMapUint32(buf_m, token.Message); err != nil {
		return ret, err
	}
	bytes_m := buf_m.Bytes()

	buf_val := new(bytes.Buffer)
	val := token.AppID + token.ChannelName + token.UidStr
	buf_val.Write([]byte(val))
	buf_val.Write(bytes_m)
	bytes_val := buf_val.Bytes()

	buf_sig := hmac.New(sha256.New, []byte(token.AppCertificate))
	buf_sig.Write(bytes_val)
	bytes_sig := buf_sig.Sum(nil)

	crc32q := crc32.MakeTable(0xedb88320)
	crc_channel_name := crc32.Checksum([]byte(token.ChannelName), crc32q)
	crc_uid := crc32.Checksum([]byte(token.UidStr), crc32q)

	buf_content := new(bytes.Buffer)
	if err := packString(buf_content, string(bytes_sig[:])); err != nil {
		return ret, err
	}
	if err := packUint32(buf_content, crc_channel_name); err != nil {
		return ret, err
	}
	if err := packUint32(buf_content, crc_uid); err != nil {
		return ret, err
	}
	if err := packString(buf_content, string(bytes_m[:])); err != nil {
		return ret, err
	}
	bytes_content := buf_content.Bytes()

	ret = version + token.AppID + base64.StdEncoding.EncodeToString(bytes_content)
	return ret, nil
}

func packUint16(w io.Writer, n uint16) error {
	return binary.Write(w, binary.LittleEndian, n)
}

func packUint32(w io.Writer, n uint32) error {
	return binary.Write(w, binary.LittleEndian, n)
}

func packString(w io.Writer, s string) error {
	err := packUint16(w, uint16(len(s)))
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(s))
	return err
}

func packMapUint32(w io.Writer, extra map[uint16]uint32) error {
	keys := []int{}
	if err := packUint16(w, uint16(len(extra))); err != nil {
		return err
	}
	for k := range extra {
		keys = append(keys, int(k))
	}
	//should sorted keys
	sort.Ints(keys)

	for _, k := range keys {
		v := extra[uint16(k)]
		if err := packUint16(w, uint16(k)); err != nil {
			return err
		}
		if err := packUint32(w, v); err != nil {
			return err
		}
	}
	return nil
}
