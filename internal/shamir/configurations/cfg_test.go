package configurations

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	FirstDecryptedKey = `data: |
    -----BEGIN RSA PRIVATE KEY-----
    MIICXAIBAAKBgQCzHNJIrdg7AIEcCt1sT4vo7nrWLCdy8tSFv1s3S6dUKoC+nvtz
    F7VhWHZNe5TTlqnJM/W59D+y8BzGd3UhG5Trsu0QP6ZBrTQ9vWcZ95CoQqxWYsMW
    FHGWKMhirHS/Oq9apNLlLUIABjIwX79fAQkzWvIhzuaZue/enzWkqVIkAQIDAQAB
    AoGAfqDItiNZtNT1clroEhPm6SX46APNXAH7aBdSKLSutt1ZqeqCAOtpC0pcUQkm
    vbRNkvbOKcpZYmtJHLMFtwNl81HA865wssC4PSkSX4lT57FxBSdX+nCkr1I5yLKQ
    SHu8SVL+Kzayc4madbFhz7M8x7KUFlPQrin0CMy0JStoR4UCQQDfPPegTEXSKakt
    p5IVXXaGgebppYxNDpVmVNILRFez2LluXNclSSmvqZS0aKXWBl7mXmqdSJnQPsdd
    eEUhH2zzAkEAzWYPtUdsbJTIhkT7zIAAD7bVYNTRAxdwtOAY5th7N+8bK1u3Brg6
    xQboThH9N2KIV+96nahEAI3Ui/dBjf3YOwJBAM2GgvDY5/tKkdBPk6YJ62KA3EhG
    OvqCr7eL5emrnRceM/JQlV9lkXSBqz/BjNST/vEXAg8WVU4tTj1WykZpeVMCQFqF
    l3iqMKOt8q4SKvLWmrBSQLn7MN5u7zIX7YGXeL6O92dldUVV3jjFMY5uqS1GlJZE
    qcJfzRB9cWX6I38Dl88CQHWYVendyVKJfNyM0SPx9W7uf/2ETHpsA0RsRTRmQJlo
    9B5GBOWU94iBd9VIiot3K1AazRbEnW67/chRhPsBKdk=
    -----END RSA PRIVATE KEY-----`
	SecondDecryptedKey = `data: |
    -----BEGIN RSA PRIVATE KEY-----
    MIICXQIBAAKBgQDF/oLT2CKigD+DVXh070lUFvl/TqU88jBQYdNGdc2B/sOqac7N
    zNPUgUQrE8iWMe3TgraDZisnohKHGwlBfKyRAOvSkXLmsIUac6AmgnUfWJW1STTu
    hCCQRN6+joJTpkcolCiAHkJXNOv5O8lMiuM8snfB+Qfg5s15XRZLIPKXOwIDAQAB
    AoGBAIzfAeyCWlouT4I9PMBORGkdE5V9ull7o959S9pVFMwKHONR9v88XfdTpdHU
    eKJMkf8sgHlLJydCzFXuaY57izc77W4D2nSd0frvp1P16mt1Yg178xf8z+bPwwpw
    gOrJY/VhsQbSCA1ojl1Kr1WHRR9mG8r4q6bQo0VVmprQtkUxAkEA5sya7jkQ+BGr
    w2NP5Ga11kf3mMwfvRj+/KGfis9dUAkdsi6eeSCitKSV9+cIQz3PP/y7geGwrUkg
    0d6uRO2v+QJBANuc7O5/rCL6eY4xGj/J1SqVtkcDDFgvIxX4x1mtpAV9PD8dOssw
    pv9a4lQuP8kd4hWFNus9iQmEIyy/ppd8NdMCQHUDqU08W5EaDHachGXteSWyCSNL
    4o+R+72ECApthgbNCGRNZIVO+Y8SnwI3/SVyUgBEF6ELFvNUTzJ4bux9gakCQDvF
    eSLscsBOZbkSmi9UUoU1nBe1t61knusC+9bnkKXB0RzoarEUu87yQHli4Q9u57Kt
    wCXEaMDDSwOJ4eCG+OsCQQCrisXn9vr7f3qd1re9ENu5qDdOSxg76AzX3bfCib71
    bbSYZQoZCIaqFsKFBC49jZeXJ4Ut/ydHJCXUy7CLB0EI
    -----END RSA PRIVATE KEY-----`
)

var (
	privateKeyFirstFilePath  string
	privateKeySecondFilePath string
	tmpDir                   string
	credFilePath             string
)

func init() {
	tmpDir = os.TempDir()
	privateKeyFirstFilePath = filepath.Join(tmpDir, "private_key_01.pem.encrypted.yaml")
	privateKeySecondFilePath = filepath.Join(tmpDir, "private_key_02.pem.encrypted.yaml")
}

func createPrivateKeyFirstFile() error {
	const base64PrivateKeyFileData = `ewoJImRhdGEiOiAiRU5DW0FFUzI1Nl9HQ00sZGF0YTpBU2REVlRVSHJLT0pGZTQrN3lhYWhTVGxn
SHp5c2lsVWlpbzZyUkxVTTBWemRxT3F2KzhpMzczOUYwR3NjamxYOXRhMmpJVTJjenlaa3A0N1h1
dzlZb3E3dEs2S1pNWnFnQWF6bk1xWDVvbWtsVVppSzQza2VPYlB6ZDJLZlZyQVMzek9tMSt2dUM1
anpUN21UWmlCalN2QkhpNkRnaDNobEc1ck1WVVlpNVJIbTNIL1dvdWNpVFJCN2ZRWUtQWnJvWkx2
WXJyUUUxTEp4L1QrNUs2MFhVTTBQdGFQZ29vdkF3NFVORGtMSlZHYit1WUFrS0NpaWRWT25JSHMv
VSt3UWx5U25zdjV4LzJWY2xKeXBxMk5NMEE2MHl1NkhTWjBFZkNLQWVLYWxldGRyT3lSNzVzR2lp
QjZzcU1YUk40UVcvN2ZDQWNpSUlqRXM2bkFoUTBXRXVvWnUrT2hSS3h0d3Vqa2cxcDExd0RKeFYv
anNWRDJ4M2ZGSUgyWmdXcTRUc1JHRG5rZFNNc0laVndrZ21BTlhGNVJpZWhUQXJ5SjFPTVRnN1BG
ZERhVDVSRFhQWGhROFFxeHVCc3lrNklTdkMzNEp2eFhvckVkVXh5QXRVSU41cmk1VUZwU2dpdnNy
NzIrZGxXV3FHKy8wNmFwcEt0T2JoTUlpenlVdVp5Q2FtMFRibkNNaFNLc3ZGTlI2YXlIZnU2QzFa
aTRuTHo1Y1pDelNkUVBOczB1Q05ldWJkR0tpbVlIRnhsUlE2Qko2SVEyM3ZZb0JCTDFNUnZuYmdN
Sm5wRFdIMStia05uYnRSVUZKanQrTHBrSDVCN0M4MDVaMXRVRWhKSGZOVitPUVhPNnlGNDN1eGpz
cEZlRHo2TUpYYUJJNHE5dDdFRGdYVWJBcUNKU1AzSmh0M1l6NFdLUmdZYkowaEM0ZmFBMmdNQTM5
TTZIMUtCb2lpa0ZCOWF6Q3NtVWFOOUNnRWV5THFUUFBrdjQzUUZ5TVR3eFBlVTlXMjBDUGNCQTZK
emQramQ5MWJiOGFNZTFWSnhKSVhqWHhlU2lud3JiZGdpV0FDOVFRQnZERUR0QUJkN0ZvVlVEZ1pw
Vk14U1F5TjYyTk5BSzhnV2RpRSswb2xMZ1ZVRFFVcy9maHNNRGFzRkxvVWZiZm8yNTk3ck93TTlI
MnZlVllXQnZSNWZTU1JKMFRxdzZmTXU4SlArU09QSW5nMWVWVnFoWlJ0aTBqWXovN1BpUUNvYUFD
Q2lLUm1Ud0svK2hqVHJtZWZHaHAxWlRDQ2VqS00vT1c3SktBSWZzd2ZBMU9KcDBFRlRNV2FYN3hj
b1NDKzR0b0J3UVIySlMvQmtwekk1eTV0bXk4bHFQTkh0Y3ZzSG9qRnNWTGpqckZwMHJSdmh6Q2xt
aW93M0FIVDE1TTgwUDRqTnVPWUd6a1Vza2V3YWdlTDNZdGd6eWtOeW5ianptcTZ1aE9BangrTnlH
WlZJVlU5U1kzOVBiNlRxa0x2YVVwNG53Ukc1RDJnNS9uOFhxZ1JERWk3dnRTLzFMbVlIZEpwU3Ju
am5oM0VrRFhZNTdEaHFhSkdRPSxpdjpPbm9ZOFR0VnhlcVVlcG9CYkRpcG5ETWQ0Z2g5UmJXVnZG
cUdVdTNoWUY4PSx0YWc6MXAyM3pYZ0dJaHpmMGdqTmpwek1idz09LHR5cGU6c3RyXSIsCgkic29w
cyI6IHsKCQkia21zIjogbnVsbCwKCQkiZ2NwX2ttcyI6IFsKCQkJewoJCQkJInJlc291cmNlX2lk
IjogInByb2plY3RzL2Rldi1tYW5hYmllLW9ubGluZS9sb2NhdGlvbnMvZ2xvYmFsL2tleVJpbmdz
L2RlcGxveW1lbnRzL2NyeXB0b0tleXMvZ2l0aHViLWFjdGlvbnMiLAoJCQkJImNyZWF0ZWRfYXQi
OiAiMjAyMi0wMS0wNVQwOTo0NToxM1oiLAoJCQkJImVuYyI6ICJDaVFBLzJsSGpVVVEzYnEyT0tB
a2JtbnhTREFvalFCYS9UQVFML1Z2dG1Xd3Zxb2tnQUFTU1FDYWJDb3R3a0NYYUN3c3JCRGtSSVdY
dy83WVpxWWNTc000OUlqQ0xPVWR5NHRmRTZyNG5lY1RiNjBwNGRyaDJWN3RrTE5RTXFPOVZiZVFU
Yk1md2JiT3dIQklNYUJJbkZRPSIKCQkJfQoJCV0sCgkJImF6dXJlX2t2IjogbnVsbCwKCQkiaGNf
dmF1bHQiOiBudWxsLAoJCSJhZ2UiOiBudWxsLAoJCSJsYXN0bW9kaWZpZWQiOiAiMjAyMi0wMS0w
NVQwOTo0NToxM1oiLAoJCSJtYWMiOiAiRU5DW0FFUzI1Nl9HQ00sZGF0YTpNNk5uRDFnVkxseXpa
aFd0L1ZxWnUxU254NGNvVS8zMlcyS0xrS1YwTGdlSG01djJldGlUVS85UnBUT2d0eElQVElzOGx1
dTVKVzFzWWVPclc5WFpPendLNjFBRmNIYVFnSXBFOFZXa3BPM1lQdTBPa1o5b3dTa2dHM3ZwVDdu
YUVJaVhWWkF1aDNRdEJyU0FJenI0UmVFS0dGSVYwWFNLcGtvK3k2Q21uR2c9LGl2OjI0aUxQelZO
Z2FWbFdrRGlpYVh6M3ZTcDh6c21sVFF0V2Jac1NzcGVDenM9LHRhZzpQVDRzdTIxNW9Oa0htaXlv
Uy9wNkZ3PT0sdHlwZTpzdHJdIiwKCQkicGdwIjogbnVsbCwKCQkidW5lbmNyeXB0ZWRfc3VmZml4
IjogIl91bmVuY3J5cHRlZCIsCgkJInZlcnNpb24iOiAiMy43LjEiCgl9Cn0=
`
	secretData, err := base64.StdEncoding.DecodeString(base64PrivateKeyFileData)
	if err != nil {
		return fmt.Errorf("failed to decode base64: %s", err)
	}
	err = os.WriteFile(privateKeyFirstFilePath, secretData, 0666)
	if err != nil {
		return fmt.Errorf("failed to write to secret file: %s", err)
	}
	return nil
}

func createPrivateKeySecondFile() error {
	const base64PrivateKeyFileData = `ewoJImRhdGEiOiAiRU5DW0FFUzI1Nl9HQ00sZGF0YTpodlQrZU1UbFVweWZ4SHZjUENjTnRwWHEr
czFFUjlINWJOSkdpSzVQWDZQc1BmWE5Ga0I3MU8xaEpFVTZUZi9xbXJ5NGV5dzE5dlE4NS82WlJj
T05CdGFLeTNMWXhQQnlrRnRIRkFCYUJZK3Y4dTdRTGwyeGY3azVSQklVZko0QWE0c3dCTWV0amsr
cUF1cTJ5RnZSNWhscE5qODdiVWRqNU9VUU5GbVJidWVBc2FWMUNnNURzSitiOG5VcFhHY2ZHcldX
TTdMWXRHaDd1Y3RwYlFqeGZwVXd3LytBMjhvWHBTdFpubmJzaVhzMmc1MFBSdTdjUXVEcnlsLzBk
NXZJSW4yYzE0eVZURWdPanNmajB0dVd2dWt2S0NqV2RzZW5pZ09CVWgzblpPd0JmK3VmR1FXTkFV
TzREeklBL2Jzdzh6NVVJMEtYdE9Pa083SExEbEhaNWZVcmcrdWk1ZHlxU2VnYnVORi9pTWx3dFVN
SDJWdHE5MXZieEhCTXBtMWZZeVhmb2V0UkVJb21YTFE4NVdnNGM4YVhFaWY0dkhCWkdrL0VXUlE2
alU4dnQ2aE1GSzB0SHNZZjVmemxFOHVBeGpwTjFLUFpwQTBOZ2hiTnpDYm5oN2lKcW1kZEpjRnB0
UTR2OHQ5YUxTbDFmYVRxcXNHdzZsUWFuWG81bU1Lb2svb3dvNmhveFJBbG9mVFN2cHBNdFJUKzRw
RTFXUVQ1YWx3a0drRFFzL0xMM2tFN1FlMjVyVDdDRGlxVHlnYUlVV1QzZEsrSFR0T0puTk05SVFY
ZjhoWVJyUmNwY2dIelRnTC9VVFpQV3VaQmd1UXpONDFibGV6MXNNTEZWb3NESXdlb1NINWVBNVMr
enlJMkNUbG4vOGV1MHlrVG9hMisyOEY4U01IcjMwU3NWRVl3VDFiQkd6K24rTUEzR25JVmhwMmsx
U3RpVjR4R1JjWGJhYVFhV01OV3BNZ3lzQTllcUszcFhqem4wYnd6TnFVK1ViWHdQKzRxZHh6MW1I
NlNuUHFqZXR4blFqTTNRalVpbVNWSkZPRHlzN2ZrQWE1ZGs0Yk5vTUo3MTVhcWNOeUcweG1yZlNZ
Qk5JaXdaaXNnYnNKM0lCRmNRQjVFNUJXdGhaY0lrMmZuUFVmMkpWRzNiaWthNVZkMnRhcVVSYU4z
RUFtaGpqMU1ESTlzRE1INWJPV2huM0V5YXNEQmdTMEQ5eDJMVmVXYitWaGpJSGFac0UwV0JuNG1j
UVRDOWFlRFlVQkR2T3ZJV1Y0RmlIRVRqcnpoYTdkU21pa3pTYjZXY3Qyb2xSUEVqU3V1aEhJSWpR
NXUwRFdNd1hGMldvWkhOTkJ5Yks3MzFtZGp2dGdZUldtMnM3Z2hodjYzTlBEWUJ3VmtIM0VWdm1s
S1RuU1N2OUJGYXlCT21lNmRrT1cwSjhMTnVjdk00aCtEV3Z0MDRKZTVQZ1dRUFRhcHJsK0l0R2ls
czg5MmhHWlRJaGVZRWo1N2l5VUhxV0NPYmxrUWdFczM5Q1RwL3dXdFZLdnQ0ZDdpUXZQTi92OW92
ajJ3S2lFUGZqVThFbnYvaUlvPSxpdjoxSWUxcmJLRklOMTc0TFdtK2diVDArRGpGc1oxUW1NbHhG
QnFkc243NlFjPSx0YWc6c2xpSndHVVJsenBZQzMrOVdCb3k1dz09LHR5cGU6c3RyXSIsCgkic29w
cyI6IHsKCQkia21zIjogbnVsbCwKCQkiZ2NwX2ttcyI6IFsKCQkJewoJCQkJInJlc291cmNlX2lk
IjogInByb2plY3RzL2Rldi1tYW5hYmllLW9ubGluZS9sb2NhdGlvbnMvZ2xvYmFsL2tleVJpbmdz
L2RlcGxveW1lbnRzL2NyeXB0b0tleXMvZ2l0aHViLWFjdGlvbnMiLAoJCQkJImNyZWF0ZWRfYXQi
OiAiMjAyMi0wMS0wNVQwOTo0NTo1MVoiLAoJCQkJImVuYyI6ICJDaVFBLzJsSGpZcFFDaCtyYXJU
cG5nYkVxaW9NRzZnblRGL29sKzhZb3FRRjl1NSsrbVlTU1FDYWJDb3R2ejdwQ0JPdFpVTHdPYTFD
eHRXR01wVkFTYzBzWFhSMGVYdE5RS3VKZUF0N284b1BOY1pXY3p2dmFLT2VUN0M2eWtoa0M1VzJv
VU4xQmlzSElRUGZUK0RpL1pvPSIKCQkJfQoJCV0sCgkJImF6dXJlX2t2IjogbnVsbCwKCQkiaGNf
dmF1bHQiOiBudWxsLAoJCSJhZ2UiOiBudWxsLAoJCSJsYXN0bW9kaWZpZWQiOiAiMjAyMi0wMS0w
NVQwOTo0NTo1MVoiLAoJCSJtYWMiOiAiRU5DW0FFUzI1Nl9HQ00sZGF0YTpXc0JEUXpKcnYyTUlX
YUs4WkRjWng4WkF4OUhyZEhxclRhZVdKNXA2Y0MyVlJsenU2ZXRNZWhOZHhKWThaV3huYVpXdlhX
NzdwUU9jN2gxM3V5a0tkbVRxQUwxY0htSFBZNXIvMVZPUmlNYlkyL1VkZlFsb2N4R0syNTZzMGxw
OE54ZzN3a2g4bnlWaUlYTmlXcU1kalVSOVBlUlNWNnlnMkkyNXhYSFpscDg9LGl2OmJiV2o5WXNN
a2RPV3N6NVczeDQrS2VoTHZnSVFDcitPNDJxVFZlNjFTcWc9LHRhZzp1WXR3RGErR3MxdGxrUlRz
N25XY3RBPT0sdHlwZTpzdHJdIiwKCQkicGdwIjogbnVsbCwKCQkidW5lbmNyeXB0ZWRfc3VmZml4
IjogIl91bmVuY3J5cHRlZCIsCgkJInZlcnNpb24iOiAiMy43LjEiCgl9Cn0=
`
	secretData, err := base64.StdEncoding.DecodeString(base64PrivateKeyFileData)
	if err != nil {
		return fmt.Errorf("failed to decode base64: %s", err)
	}
	err = os.WriteFile(privateKeySecondFilePath, secretData, 0666)
	if err != nil {
		return fmt.Errorf("failed to write to secret file: %s", err)
	}
	return nil
}

func TestLoadPrivateKeysWithSopsFormat(t *testing.T) {
	t.Parallel()
	_, filename, _, _ := runtime.Caller(0)

	t.Run("load private keys not found", func(tt *testing.T) {
		tt.Parallel()
		keys, primaryKeyID, err := loadPrivateKeysWithSopsFormat(filepath.Join(filepath.Dir(filename), "test_notfound_*.pem"), "test_notfound_1.pem", mockDecrypt)
		assert.Error(tt, err)
		assert.Equal(tt, "", primaryKeyID)
		assert.Nil(tt, keys)
	})

	t.Run("load private keys", func(tt *testing.T) {
		tt.Parallel()

		err := createPrivateKeyFirstFile()
		assert.Nil(tt, err)

		err = createPrivateKeySecondFile()
		assert.Nil(tt, err)

		privateKeys, primaryKeyID, err := loadPrivateKeysWithSopsFormat(fmt.Sprintf("%s/*.pem.encrypted.yaml", tmpDir), "", mockDecrypt)
		assert.Nil(tt, err)
		assert.Equal(tt, "", primaryKeyID)
		assert.NotNil(tt, privateKeys)
	})
}
func mockDecrypt(path, format string) ([]byte, error) {
	if format != "yaml" {
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
	if path == privateKeyFirstFilePath {
		return []byte(FirstDecryptedKey), nil
	}
	if path == privateKeySecondFilePath {
		return []byte(SecondDecryptedKey), nil
	}
	return nil, fmt.Errorf("can't decrypt this file: %s", path)
}
