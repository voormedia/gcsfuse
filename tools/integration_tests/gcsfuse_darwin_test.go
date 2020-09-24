// Copyright 2015 Google Inc. All Rights Reserved.
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

// A collection of tests for a file system where we do not attempt to write to
// the file system at all. Rather we set up contents in a GCS bucket out of
// band, wait for them to be available, and then read them via the file system.

package integration_test

import (
	"math"
	"syscall"

	"github.com/simonwahlstrom/gcsfuse/internal/canned"
	. "github.com/jacobsa/ogletest"
)

////////////////////////////////////////////////////////////////////////
// Helpers
////////////////////////////////////////////////////////////////////////

func convertStatfsString(in []int8) (s string) {
	var tmp []byte
	for _, v := range in {
		if v == 0 {
			break
		}

		tmp = append(tmp, byte(v))
	}

	s = string(tmp)
	return
}

////////////////////////////////////////////////////////////////////////
// Tests
////////////////////////////////////////////////////////////////////////

func (t *GcsfuseTest) Statfs() {
	var err error
	var stat syscall.Statfs_t

	// Mount.
	args := []string{canned.FakeBucketName, t.dir}

	err = t.runGcsfuse(args)
	AssertEq(nil, err)
	defer unmount(t.dir)

	// Stat the file system.
	err = syscall.Statfs(t.dir, &stat)
	AssertEq(nil, err)

	// The FS should show a reasonable number of bytes available, so that e.g.
	// the Finder doesn't refuse to copy files into it.
	AssertLe(stat.Blocks, math.MaxUint64/uint64(stat.Bsize))
	ExpectGe(uint64(stat.Bsize)*stat.Blocks, 1<<50)
	ExpectEq(stat.Blocks, stat.Bfree)
	ExpectEq(stat.Bfree, stat.Bavail)

	// Similarly with inodes.
	ExpectGe(stat.Files, 1<<40)
	ExpectEq(stat.Files, stat.Ffree)

	// The recommended IO size should not be pitiful.
	ExpectEq(1<<20, stat.Iosize)

	// The file system name should be the bucket's name.
	ExpectEq(canned.FakeBucketName, convertStatfsString(stat.Mntfromname[:]))
}
