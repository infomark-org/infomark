// InfoMark - a platform for managing courses with
//            distributing exercise sheets and testing exercise submissions
// Copyright (C) 2019  ComputerGraphics Tuebingen
// Authors: Patrick Wieschollek
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package symbol

type key int

// to replace
//   context.WithValue(ctx, "course", course)
// and
//   r.Context().Value(symbol.CtxKeyCourse)
// TODO(): create a shared context-key package
const (
	CtxKeyAccessClaims key = iota // must be 0 to work with the auth-package
	CtxKeyGroup        key = iota
	CtxKeyMaterial     key = iota
	CtxKeyCourse       key = iota
	CtxKeyCourseRole   key = iota
	CtxKeyUser         key = iota
	CtxKeyTask         key = iota
	CtxKeySubmission   key = iota
	CtxKeySheet        key = iota
	CtxKeyGrade        key = iota
	// ...
)

type testingState int

// these are states of a submission life-cycle
const (
	TestingStateEnqueue  testingState = 0 // submission has arrived
	TestingStateRunning  testingState = 1 // submission test is running
	TestingStateFinished testingState = 2 // submission test has finished
)

type TestingResult int64

const (
	TestingResultSuccess TestingResult = 0 // docker was able to test (but test could have failed)
	TestingResultFailed  TestingResult = 1 // something went wrong with docker

)

func (t TestingResult) AsInt64() int64 {
	if t == TestingResultSuccess {
		return 0
	}
	return 1
}
