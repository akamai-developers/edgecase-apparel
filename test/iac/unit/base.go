package unit

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// apl
	aplId         string = "default/apl"
	aplLint       bool   = true
	aplName       string = "apl"
	aplRepo       string = "https://linode.github.io/apl-core"
	aplStatus     string = "deployed"
	k8sProviderId string = "525256"

	// dns
	ccaRecordId string = "123456"
	domain      string = "edgecase-apparel.app"
	domainId    string = "251816"

	lkeId    string = "123456"
	lkeLabel string = "edgecase-apparel-main"

	linodeToken  string = "55772f00-85e0-11f1-85a5-b3f1d391c9f3"
	platformName string = "edgecase-apparel-platform"

	objPrefix string = "ec-apl-"
	accessKey string = "2ab3b96e8ad111f197ae0ba8880fda5c"
	secretKey string = "372fbf768ad111f180be134ebd0c8ed838da6f9c8ad111f18"
)

var objBuckets = map[string]any{
	"loki":   objPrefix + "loki",
	"cnpg":   objPrefix + "cnpg",
	"harbor": objPrefix + "harbor",
	"thanos": objPrefix + "thanos",
	"tempo":  objPrefix + "tempo",
	"gitea":  objPrefix + "gitea",
}

type TestStack struct {
	Name    string
	Project string
	Slug    string
	Stack   string
}

func (t *TestStack) Init(name string) {
	n := strings.ToUpper(name)
	p := fmt.Sprintf("ECA_%s_PROJECT", n)
	s := fmt.Sprintf("ECA_%s_STACK", n)

	proj := os.Getenv(p)
	stack := os.Getenv(s)

	t.Name = n
	t.Project = proj
	t.Slug = filepath.Join("organization", proj, stack)
	t.Stack = stack
}

func NewTestStack(name string) TestStack {
	ts := new(TestStack)
	n := strings.ToUpper(name)
	p := fmt.Sprintf("ECA_%s_PROJECT", n)
	s := fmt.Sprintf("ECA_%s_STACK", n)

	proj := os.Getenv(p)
	stack := os.Getenv(s)

	ts.Name = n
	ts.Project = proj
	ts.Slug = filepath.Join("organization", proj, stack)
	ts.Stack = stack

	return *ts
}

var kubeconfig = `
CmFwaVZlcnNpb246IHYxCmtpbmQ6IENvbmZpZwpwcmVmZXJlbmNlczoge30KCmNsdXN0ZXJzOgot
IGNsdXN0ZXI6CiAgICBjZXJ0aWZpY2F0ZS1hdXRob3JpdHktZGF0YTogTFMwdExTMUNSVWRKVGlC
RFJWSlVTVVpKUTBGVVJTMHRMUzB0Q2sxSlNVUkpha05EUVdkeFowRjNTVUpCWjBsVldsVkhLMDlX
ZEZwUGJIaDNkbVZzV1RoSkswODRha1kzYWtOemQwUlJXVXBMYjFwSmFIWmpUa0ZSUlV3S1FsRkJk
MFpVUlZSTlFrVkhRVEZWUlVGM2QwdGhNMVpwV2xoS2RWcFlVbXhqZWtGbFJuY3dlVTVxUVROTmVr
VjVUV3BSTWsxRVVtRkdkekI1VG5wQk13cE5la1Y1VFdwUk1rMUVVbUZOUWxWNFJYcEJVa0puVGxa
Q1FVMU5RMjEwTVZsdFZubGliVll3V2xoTmQyZG5SV2xOUVRCSFExTnhSMU5KWWpORVVVVkNDa0ZS
VlVGQk5FbENSSGRCZDJkblJVdEJiMGxDUVZGRGFXWnFXRFJqUkhCbVJUTnRkWFV6UzBWaU1HWm1V
a2c0WWxKS1RVSjJiVWhMTTBGUlluQXlaMHNLUzNwSU16SlhPV2g0TUhKV1ZFdFpXa2RFZGtaV1Zs
WXpPR2d2U1dzNFZHSnBhblJ3VmtkbVQwRTVaMmhaTVdSNk9HbDRLMHAyVDFkNlpUazNUVFp2THdw
cGFVdDVhMmhtYTJWRVlsSjFTME5MUTFsQmRUVlJhR05WZVhGMGFVNWlWalJOZDFOS1QwTnBZMnBz
UWpKNlVHSXhTRmx0T0dsRWVWQkdhRGgzUVd3MENrWkVMMkZQVFZWa1ZqVXdjVXB6UVVKSmFGbEVS
bGhrYVZoMlVHWnJVeXRCVkdwalFqSkVlVE0xTUM5R2VEUkNORFEyZFV3MlJqYzNVMHRFWkM5c1dU
Y0tNMUl6V0dsMUsxRjJSMmxaWlRCV01rVjBjVXM1VFdaS2MxSkJlVTlNZUhWdFJ6VlJUSEZPVVVJ
MFMyWmxWWGwyWTBvNVRGTm9jbU5qUkVKNlJUVndZZ292THl0NGFWcDNSM1JNTjNGdE4ycEZURnBI
YjJ3MWRrTlJNRkZMUmpRckswMUlSbTF4ZFdvMGJHdzROVUZuVFVKQlFVZHFZV3BDYjAxQ01FZEJN
VlZrQ2tSblVWZENRbFJuTVZoTWVVbDRhMkZYU2xaYVlXOVlRUzlTVEVSWE5GcDFSR3BCWmtKblRs
WklVMDFGUjBSQlYyZENWR2N4V0V4NVNYaHJZVmRLVmxvS1lXOVlRUzlTVEVSWE5GcDFSR3BCVUVK
blRsWklVazFDUVdZNFJVSlVRVVJCVVVndlRVSlZSMEV4VldSRlVWRlBUVUY1UTBOdGRERlpiVlo1
WW0xV01BcGFXRTEzUkZGWlNrdHZXa2xvZG1OT1FWRkZURUpSUVVSblowVkNRVWxuUlRWc0t6Rnhk
RzFUTUd0alNsZHpWamRpTlRsV2NuSmpTMUp1YWpKemNteDZDbnBPT0cxaldXRnpVRW95YmtKUk0w
dHJhMFZZYldwSlpsVXJaa2d2TlhsQmRFdFdZbVZMZEhGT2JuUnJiMUkwYjJ4Q1JuSlRaMVJuT1Va
UFdGSlRNRkVLZDFGS0szaHNVRGw1TTBSRGRHOXZOWFJ4ZEVGU2ExaHFWRkJCVEV0QlJITjFjRkUz
YWpZMWVWQkpMM0pwYjNBM1VqRm1hMXBzUW1jMVZVRjJRazVHTVFwVVJGWkVRVmRxYmpsSWJsa3dO
RVZ6TVV0WmJ6VXdZV1IyZVhkbEswSlBlbUZYS3l0V1dFMW1XRE5MWldaVlpFUlFhekpQUlhRNE9X
aEhaRkZISzBGUkNuZG9kMnQ2U2xOamFITjFiMmRWUkZORlJucEhNamQwZG5Ca05TOVRUbll6WTJW
U01FVktkRWhXTDBWTWNUaHhTVWRRYWtacVFYQlJkWGRMVjNJd2FUWUtXbEZOTlRZd05GcE5WVWN5
VkZKb1dIZDJjVkZCZGxsQmJXOU9aRVZWZFRkNmNFWTFTbkJCVkZSTlZIUkpRVGN2TjFWelBRb3RM
UzB0TFVWT1JDQkRSVkpVU1VaSlEwRlVSUzB0TFMwdENnPT0KICAgIHNlcnZlcjogaHR0cHM6Ly9j
ODg5OTFhNC04ZDMzLTExZjEtYjEzMy01ZjA5NTU4ZWI5ZDEudXMtbWlhLTEtZ3cubGlub2RlbGtl
Lm5ldDo0NDMKICBuYW1lOiBsa2UxMjM0NQoKdXNlcnM6Ci0gbmFtZTogbGtlMTIzNDUtYWRtaW4K
ICB1c2VyOgogICAgYXMtdXNlci1leHRyYToge30KICAgIHRva2VuOiBleUpoYkdjaU9pSlNVekkx
TmlJc0ltdHBaQ0k2SW5wTU1WVktWRmxrUjA4dE1EWm1WSGd5YTNORlVEQlpaVkJ0TVRaeVVGcEVk
MWRwUVZCYWIwMDJjMGtpZlEuZXlKaGRXUWlPbHNpYUhSMGNITTZMeTlyZFdKbGNtNWxkR1Z6TG1S
bFptRjFiSFF1YzNaakxtTnNkWE4wWlhJdWJHOWpZV3dpWFN3aVpYaHdJam94TnpnMU5UUXlNRGd4
TENKcFlYUWlPakUzT0RVMU16ZzBPREVzSW1semN5STZJbWgwZEhCek9pOHZhM1ZpWlhKdVpYUmxj
eTVrWldaaGRXeDBMbk4yWXk1amJIVnpkR1Z5TG14dlkyRnNJaXdpYW5ScElqb2lNbUZrWXprMU9U
QXRZelEyTkMwME0yTXpMVGcwTlRFdE1EUTRZMkUwTm1aa056RTJJaXdpYTNWaVpYSnVaWFJsY3k1
cGJ5STZleUp1WVcxbGMzQmhZMlVpT2lKa1pXWmhkV3gwSWl3aWMyVnlkbWxqWldGalkyOTFiblFp
T25zaWJtRnRaU0k2SW14clpTMHhNak0wTlNJc0luVnBaQ0k2SW1SaE16azFPR1kzTFRWbFl6QXRO
R1ZqT1MxaU5HWmhMV05sTVdZMVltUXpZakpsTlNKOWZTd2libUptSWpveE56ZzFOVE00TkRneExD
SnpkV0lpT2lKemVYTjBaVzA2YzJWeWRtbGpaV0ZqWTI5MWJuUTZaR1ZtWVhWc2REcHNhMlV0TVRJ
ek5EVWlmUS5qVEhFeWpLM3NfNGF1cTdtX0s5LWh4bnBfclhPUWRWWGwyNHd2Q25LUzRJMWJTVU5H
TkEwS3pqVG50Wm5pLWpCWEZLdUw2b2VKdXJuaThZWWFzTEpndGtQcTNGWDlUSDM4a054QUFYOXNn
ZERNZEpxYVZOU1d6YWxxUTdmWXM2QXAwY0NUQXRtdFZtR0ZpaEctajZwaXlwM3pNdlNpaUZwRnlK
R0xiTkMzLWRuYWtfdGhMNGtzaHpUWUhLSWRyb2N2M1Jzd1hHZ05keUl5Z3VFQWRTcmMtTWhOWHpW
WDBKcThHeGcwYW5sVXo2bGs1OHZUbTh5QnJDX2lWYUJlS0xXTmVoeHdQRHd4VWxZMDNvSDVyMjc4
Z2dNV2pMdkNFRmExNjBfSExwcktNc3ZzWGFCbDlSQjNFOXQyNUVYNmdPTlZkVnpiTXBDSjJJS2Zq
YV90WXdIVVEKCmNvbnRleHRzOgotIGNvbnRleHQ6CiAgICBjbHVzdGVyOiBsa2UxMjM0NQogICAg
bmFtZXNwYWNlOiBkZWZhdWx0CiAgICB1c2VyOiBsa2UxMjM0NS1hZG1pbgogIG5hbWU6IGxrZTEy
MzQ1LWN0eAoKY3VycmVudC1jb250ZXh0OiBsa2UxMjM0NS1jdHgK`
