// InfoMark - a platform for managing courses with
//            distributing exercise sheets and testing exercise submissions
// Copyright (C) 2019 ComputerGraphics Tuebingen
//               2020-present InfoMark.org
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

import (
	"testing"

	"github.com/coreos/go-semver/semver"
	"github.com/franela/goblin"
)

func TestVersion(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("Version value", func() {

		// The exported Version must be assembled from the individual component
		// variables so that they cannot drift apart.
		g.It("Should be assembled from its components", func() {
			g.Assert(Version.Major).Equal(VersionMajor)
			g.Assert(Version.Minor).Equal(VersionMinor)
			g.Assert(Version.Patch).Equal(VersionPatch)
			g.Assert(string(Version.PreRelease)).Equal(VersionPre)
		})

		// semver ordering must place the current prerelease strictly below the
		// first stable release, which is the invariant a prerelease encodes.
		g.It("Should sort below the first stable release", func() {
			firstStable := semver.Version{Major: 1, Minor: 0, Patch: 0}
			g.Assert(Version.LessThan(firstStable)).IsTrue()
		})

		// A prerelease of a version sorts below the same version without the
		// prerelease suffix; verify the general semver rule holds for our value.
		g.It("Should sort below its own stable counterpart", func() {
			stableCounterpart := semver.Version{
				Major: VersionMajor,
				Minor: VersionMinor,
				Patch: VersionPatch,
			}
			g.Assert(Version.LessThan(stableCounterpart)).IsTrue()
		})

		// A newer patch of the same prerelease line must sort above the current
		// version, confirming Compare orders patch numbers as expected.
		g.It("Should sort below a higher patch on the same line", func() {
			higherPatch := semver.Version{
				Major:      VersionMajor,
				Minor:      VersionMinor,
				Patch:      VersionPatch + 1,
				PreRelease: semver.PreRelease(VersionPre),
			}
			g.Assert(Version.LessThan(higherPatch)).IsTrue()
		})
	})

	g.Describe("TestingResult.AsInt64", func() {

		// The success result must map to zero and the failure result to one,
		// mirroring the process return-code convention documented in the source.
		g.It("Should map success to zero and failure to one", func() {
			g.Assert(TestingResultSuccess.AsInt64()).Equal(int64(0))
			g.Assert(TestingResultFailed.AsInt64()).Equal(int64(1))
		})
	})

	g.Describe("Context keys", func() {

		// The access-claims key must be zero, which the auth package relies on,
		// and the remaining keys must be distinct successive values.
		g.It("Should start at zero for the access-claims key", func() {
			g.Assert(int(CtxKeyAccessClaims)).Equal(0)
			g.Assert(CtxKeyGroup == CtxKeyAccessClaims).IsFalse()
			g.Assert(CtxKeyCourse == CtxKeyGroup).IsFalse()
		})
	})
}
