package util

import (
	"crypto/sha1"
	"fmt"
	"io"
	"os"
)

func GetSha1Signature(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha1.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	sha := fmt.Sprintf("%x", hash.Sum(nil))
	return sha, nil
}

func GenerateBundleSignature(bundleDirectory string, bundlePath string) (string, string, error) {
	packagePath, err := CompressDirectory(bundleDirectory, bundlePath)
	if err != nil {
		return "", "", err
	}

	packageSignature, err := GetSha1Signature(packagePath)
	if err != nil {
		return "", "", err
	}

	return packagePath, packageSignature, nil
}
