/*
 * Copyright 2018 the original author or authors.
 *
 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at
 *  
 *        http://www.apache.org/licenses/LICENSE-2.0
 *  
 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package kubectl

import (
	"github.com/projectriff/riff/riff-cli/pkg/osutils"
	"time"
)

var EXEC_FOR_STRING = ExecForString
var EXEC_FOR_BYTES = ExecForBytes

func ExecForString(cmdArgs []string) (string, error) {
	out, err := EXEC_FOR_BYTES(cmdArgs)
	return string(out), err
}

func ExecForBytes(cmdArgs []string) ([]byte, error) {
	return osutils.Exec("kubectl", cmdArgs, 20*time.Second)
}
