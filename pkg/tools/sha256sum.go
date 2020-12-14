// Copyright 2019 Thorsten Kukuk
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tools

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

func Sha256sum_b(buffer string) (result string, err error) {
	hash := sha256.New()
	hash.Write([]byte(buffer))
	result = hex.EncodeToString(hash.Sum(nil))
	return result, nil
}

func Sha256sum_f(filePath string) (result string, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return "", err
	}

	result = hex.EncodeToString(hash.Sum(nil))
	return result, nil
}
