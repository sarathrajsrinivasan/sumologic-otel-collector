package main

import (
	"bytes"
	"io/fs"
	"path"
	"testing"
	"testing/fstest"
)

func TestDeleteTagAction(t *testing.T) {
	tests := []struct {
		Name                   string
		Conf                   fs.FS
		Flags                  []string
		ExpConfDWrite          []byte
		ExpConfDOverridesWrite []byte
		ExpErr                 bool
	}{
		{
			Name:  "doc missing",
			Conf:  fstest.MapFS{},
			Flags: []string{"--delete-tag", "foo"},
		},
		{
			Name: "doc missing, overrides present",
			Conf: fstest.MapFS{
				path.Join(ConfDotD, ConfDOverrides): &fstest.MapFile{
					Data: []byte("extensions:\n  sumologic:\n    collector_fields:\n      foo: bar\n"),
				},
			},
			Flags: []string{"--delete-tag", "foo"},
		},
		{
			Name: "tag deleted",
			Conf: fstest.MapFS{
				path.Join(ConfDotD, ConfDSettings): &fstest.MapFile{
					Data: []byte("extensions:\n  sumologic:\n    collector_fields:\n      foo: bar\n      bar: baz\n"),
				},
			},
			Flags:         []string{"--delete-tag", "foo"},
			ExpConfDWrite: []byte("extensions:\n  sumologic:\n    collector_fields:\n      bar: baz\n"),
		},
		{
			Name: "delete override",
			Conf: fstest.MapFS{
				path.Join(ConfDotD, ConfDOverrides): &fstest.MapFile{
					Data: []byte("extensions:\n  sumologic:\n    collector_fields:\n      foo: bar\n      bar: baz\n"),
				},
			},
			Flags:                  []string{"--override", "--delete-tag", "foo"},
			ExpConfDOverridesWrite: []byte("extensions:\n  sumologic:\n    collector_fields:\n      bar: baz\n"),
		},
		{
			Name: "delete settings with override",
			Conf: fstest.MapFS{
				path.Join(ConfDotD, ConfDSettings): &fstest.MapFile{
					Data: []byte("extensions:\n  sumologic:\n    collector_fields:\n      foo: bar\n      bar: baz\n"),
				},
			},
			Flags:         []string{"--override", "--delete-tag", "foo"},
			ExpConfDWrite: []byte("extensions:\n  sumologic:\n    collector_fields:\n      bar: baz\n"),
		},
		{
			Name: "delete both settings and overrides",
			Conf: fstest.MapFS{
				path.Join(ConfDotD, ConfDSettings): &fstest.MapFile{
					Data: []byte("extensions:\n  sumologic:\n    collector_fields:\n      foo: bar\n      bar: baz\n"),
				},
				path.Join(ConfDotD, ConfDOverrides): &fstest.MapFile{
					Data: []byte("extensions:\n  sumologic:\n    collector_fields:\n      foo: bar\n      bar: baz\n"),
				},
			},
			Flags:                  []string{"--override", "--delete-tag", "foo"},
			ExpConfDWrite:          []byte("extensions:\n  sumologic:\n    collector_fields:\n      bar: baz\n"),
			ExpConfDOverridesWrite: []byte("extensions:\n  sumologic:\n    collector_fields:\n      bar: baz\n"),
		},
		{
			Name: "error when key exists in a user-controlled file and override is used",
			Conf: fstest.MapFS{
				path.Join(ConfDotD, ConfDSettings): &fstest.MapFile{
					Data: []byte("extensions:\n  sumologic:\n    collector_fields:\n      foo: bar\n      bar: baz\n"),
				},
				path.Join(ConfDotD, "user-settings.yaml"): &fstest.MapFile{
					Data: []byte("extensions:\n  sumologic:\n    collector_fields:\n      foo: bar\n      bar: baz\n"),
				},
			},
			Flags:  []string{"--override", "--delete-tag", "foo"},
			ExpErr: true,
		},
		{
			Name: "no error when key exists in a user-controlled file and override is not used",
			Conf: fstest.MapFS{
				path.Join(ConfDotD, ConfDSettings): &fstest.MapFile{
					Data: []byte("extensions:\n  sumologic:\n    collector_fields:\n      foo: bar\n      bar: baz\n"),
				},
				path.Join(ConfDotD, "user-settings.yaml"): &fstest.MapFile{
					Data: []byte("extensions:\n  sumologic:\n    collector_fields:\n      foo: bar\n      bar: baz\n"),
				},
			},
			Flags:         []string{"--delete-tag", "foo"},
			ExpConfDWrite: []byte("extensions:\n  sumologic:\n    collector_fields:\n      bar: baz\n"),
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			var settingsWriter, overridesWriter func([]byte) (int, error)

			if test.ExpConfDWrite != nil {
				settingsWriter = newTestWriter(test.ExpConfDWrite).Write
			} else {
				settingsWriter = errWriter{}.Write
			}

			if test.ExpConfDOverridesWrite != nil {
				overridesWriter = newTestWriter(test.ExpConfDOverridesWrite).Write
			} else {
				overridesWriter = errWriter{}.Write
			}

			stdoutBuf := new(bytes.Buffer)
			stderrBuf := new(bytes.Buffer)

			actionContext := makeTestActionContext(t,
				test.Conf,
				test.Flags,
				stdoutBuf,
				stderrBuf,
				settingsWriter,
				overridesWriter,
			)

			err := DeleteTagAction(actionContext)
			if test.ExpErr && err == nil {
				t.Fatal("expected non-nil error")
			}
			if !test.ExpErr && err != nil {
				t.Fatal(err)
			}
		})
	}
}
