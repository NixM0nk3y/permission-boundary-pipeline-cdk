package util

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
)

func CalculateQualifier(Tenant, Application string) string {
	// generate semi consistant CDK indentifier
	// used to seperate this applications roles + permissions
	hash := md5.New()
	hash.Write([]byte(fmt.Sprintf("%s%s", Tenant, Application)))
	hexid := hash.Sum(nil)[0:5]
	return hex.EncodeToString(hexid)
}
