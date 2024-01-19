//go:build rocksdb
// +build rocksdb

package opendb

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/linxGnu/grocksdb"
	"github.com/stretchr/testify/require"
)

type mockAppOptions struct {
	opts map[string]interface{}
}

func newMockAppOptions(opts map[string]interface{}) *mockAppOptions {
	return &mockAppOptions{
		opts: opts,
	}
}

func (m *mockAppOptions) Get(key string) interface{} {
	return m.opts[key]
}

func TestOpenRocksdb(t *testing.T) {
	t.Run("db already exists", func(t *testing.T) {
		defaultOpts := newDefaultOptions()

		for _, tc := range []struct {
			desc                  string
			mockAppOptions        *mockAppOptions
			maxOpenFiles          int
			maxFileOpeningThreads int
			writeBufferSize       uint64
			numLevels             int
		}{
			{
				desc:                  "default options",
				mockAppOptions:        newMockAppOptions(map[string]interface{}{}),
				maxOpenFiles:          defaultOpts.GetMaxOpenFiles(),
				maxFileOpeningThreads: defaultOpts.GetMaxFileOpeningThreads(),
				writeBufferSize:       defaultOpts.GetWriteBufferSize(),
				numLevels:             defaultOpts.GetNumLevels(),
			},
			{
				desc: "change 2 options",
				mockAppOptions: newMockAppOptions(map[string]interface{}{
					maxOpenFilesDBOptName:    999,
					writeBufferSizeCFOptName: 999_999,
				}),
				maxOpenFiles:          999,
				maxFileOpeningThreads: defaultOpts.GetMaxFileOpeningThreads(),
				writeBufferSize:       999_999,
				numLevels:             defaultOpts.GetNumLevels(),
			},
			{
				desc: "change 4 options",
				mockAppOptions: newMockAppOptions(map[string]interface{}{
					maxOpenFilesDBOptName:          999,
					maxFileOpeningThreadsDBOptName: 9,
					writeBufferSizeCFOptName:       999_999,
					numLevelsCFOptName:             9,
				}),
				maxOpenFiles:          999,
				maxFileOpeningThreads: 9,
				writeBufferSize:       999_999,
				numLevels:             9,
			},
		} {
			t.Run(tc.desc, func(t *testing.T) {
				dir, err := os.MkdirTemp("", "rocksdb")
				require.NoError(t, err)
				defer func() {
					err := os.RemoveAll(dir)
					require.NoError(t, err)
				}()

				db, err := openRocksdb(dir, tc.mockAppOptions)
				require.NoError(t, err)
				require.NoError(t, db.Close())

				dbOpts, cfOpts, err := LoadLatestOptions(filepath.Join(dir, "application.db"))
				require.NoError(t, err)
				require.Equal(t, tc.maxOpenFiles, dbOpts.GetMaxOpenFiles())
				require.Equal(t, tc.maxFileOpeningThreads, dbOpts.GetMaxFileOpeningThreads())
				require.Equal(t, tc.writeBufferSize, cfOpts.GetWriteBufferSize())
				require.Equal(t, tc.numLevels, cfOpts.GetNumLevels())
			})
		}
	})

	t.Run("db doesn't exist yet", func(t *testing.T) {
		defaultOpts := newDefaultOptions()

		dir, err := os.MkdirTemp("", "rocksdb")
		require.NoError(t, err)
		defer func() {
			err := os.RemoveAll(dir)
			require.NoError(t, err)
		}()

		mockAppOpts := newMockAppOptions(map[string]interface{}{})
		db, err := openRocksdb(dir, mockAppOpts)
		require.NoError(t, err)
		require.NoError(t, db.Close())

		dbOpts, cfOpts, err := LoadLatestOptions(filepath.Join(dir, "application.db"))
		require.NoError(t, err)
		require.Equal(t, defaultOpts.GetMaxOpenFiles(), dbOpts.GetMaxOpenFiles())
		require.Equal(t, defaultOpts.GetMaxFileOpeningThreads(), dbOpts.GetMaxFileOpeningThreads())
		require.Equal(t, defaultOpts.GetWriteBufferSize(), cfOpts.GetWriteBufferSize())
		require.Equal(t, defaultOpts.GetNumLevels(), cfOpts.GetNumLevels())
	})
}

func TestLoadLatestOptions(t *testing.T) {
	t.Run("db already exists", func(t *testing.T) {
		defaultOpts := newDefaultOptions()

		const testCasesNum = 3
		dbOptsList := make([]*grocksdb.Options, testCasesNum)
		cfOptsList := make([]*grocksdb.Options, testCasesNum)

		dbOptsList[0] = newDefaultOptions()
		cfOptsList[0] = newDefaultOptions()

		dbOptsList[1] = newDefaultOptions()
		dbOptsList[1].SetMaxOpenFiles(999)
		cfOptsList[1] = newDefaultOptions()
		cfOptsList[1].SetWriteBufferSize(999_999)

		dbOptsList[2] = newDefaultOptions()
		dbOptsList[2].SetMaxOpenFiles(999)
		dbOptsList[2].SetMaxFileOpeningThreads(9)
		cfOptsList[2] = newDefaultOptions()
		cfOptsList[2].SetWriteBufferSize(999_999)
		cfOptsList[2].SetNumLevels(9)

		for _, tc := range []struct {
			desc                  string
			dbOpts                *grocksdb.Options
			cfOpts                *grocksdb.Options
			maxOpenFiles          int
			maxFileOpeningThreads int
			writeBufferSize       uint64
			numLevels             int
		}{
			{
				desc:                  "default options",
				dbOpts:                dbOptsList[0],
				cfOpts:                cfOptsList[0],
				maxOpenFiles:          defaultOpts.GetMaxOpenFiles(),
				maxFileOpeningThreads: defaultOpts.GetMaxFileOpeningThreads(),
				writeBufferSize:       defaultOpts.GetWriteBufferSize(),
				numLevels:             defaultOpts.GetNumLevels(),
			},
			{
				desc:                  "change 2 options",
				dbOpts:                dbOptsList[1],
				cfOpts:                cfOptsList[1],
				maxOpenFiles:          999,
				maxFileOpeningThreads: defaultOpts.GetMaxFileOpeningThreads(),
				writeBufferSize:       999_999,
				numLevels:             defaultOpts.GetNumLevels(),
			},
			{
				desc:                  "change 4 options",
				dbOpts:                dbOptsList[2],
				cfOpts:                cfOptsList[2],
				maxOpenFiles:          999,
				maxFileOpeningThreads: 9,
				writeBufferSize:       999_999,
				numLevels:             9,
			},
		} {
			t.Run(tc.desc, func(t *testing.T) {
				name := "application"
				dir, err := os.MkdirTemp("", "rocksdb")
				require.NoError(t, err)
				defer func() {
					err := os.RemoveAll(dir)
					require.NoError(t, err)
				}()

				db, err := newRocksDBWithOptions(name, dir, tc.dbOpts, tc.cfOpts, grocksdb.NewDefaultReadOptions(), true, defaultReportMetricsIntervalSecs)
				require.NoError(t, err)
				require.NoError(t, db.Close())

				dbOpts, cfOpts, err := LoadLatestOptions(filepath.Join(dir, "application.db"))
				require.NoError(t, err)
				require.Equal(t, tc.maxOpenFiles, dbOpts.GetMaxOpenFiles())
				require.Equal(t, tc.maxFileOpeningThreads, dbOpts.GetMaxFileOpeningThreads())
				require.Equal(t, tc.writeBufferSize, cfOpts.GetWriteBufferSize())
				require.Equal(t, tc.numLevels, cfOpts.GetNumLevels())
			})
		}
	})

	t.Run("db doesn't exist yet", func(t *testing.T) {
		defaultOpts := newDefaultOptions()

		dir, err := os.MkdirTemp("", "rocksdb")
		require.NoError(t, err)
		defer func() {
			err := os.RemoveAll(dir)
			require.NoError(t, err)
		}()

		dbOpts, cfOpts, err := LoadLatestOptions(filepath.Join(dir, "application.db"))
		require.NoError(t, err)
		require.Equal(t, defaultOpts.GetMaxOpenFiles(), dbOpts.GetMaxOpenFiles())
		require.Equal(t, defaultOpts.GetMaxFileOpeningThreads(), dbOpts.GetMaxFileOpeningThreads())
		require.Equal(t, defaultOpts.GetWriteBufferSize(), cfOpts.GetWriteBufferSize())
		require.Equal(t, defaultOpts.GetNumLevels(), cfOpts.GetNumLevels())
	})
}

func TestOverrideDBOpts(t *testing.T) {
	defaultOpts := newDefaultOptions()

	for _, tc := range []struct {
		desc                  string
		mockAppOptions        *mockAppOptions
		maxOpenFiles          int
		maxFileOpeningThreads int
	}{
		{
			desc:                  "override nothing",
			mockAppOptions:        newMockAppOptions(map[string]interface{}{}),
			maxOpenFiles:          defaultOpts.GetMaxOpenFiles(),
			maxFileOpeningThreads: defaultOpts.GetMaxFileOpeningThreads(),
		},
		{
			desc: "override max-open-files",
			mockAppOptions: newMockAppOptions(map[string]interface{}{
				maxOpenFilesDBOptName: 999,
			}),
			maxOpenFiles:          999,
			maxFileOpeningThreads: defaultOpts.GetMaxFileOpeningThreads(),
		},
		{
			desc: "override max-file-opening-threads",
			mockAppOptions: newMockAppOptions(map[string]interface{}{
				maxFileOpeningThreadsDBOptName: 9,
			}),
			maxOpenFiles:          defaultOpts.GetMaxOpenFiles(),
			maxFileOpeningThreads: 9,
		},
		{
			desc: "override max-open-files and max-file-opening-threads",
			mockAppOptions: newMockAppOptions(map[string]interface{}{
				maxOpenFilesDBOptName:          999,
				maxFileOpeningThreadsDBOptName: 9,
			}),
			maxOpenFiles:          999,
			maxFileOpeningThreads: 9,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			dbOpts := newDefaultOptions()
			dbOpts = overrideDBOpts(dbOpts, tc.mockAppOptions)

			require.Equal(t, tc.maxOpenFiles, dbOpts.GetMaxOpenFiles())
			require.Equal(t, tc.maxFileOpeningThreads, dbOpts.GetMaxFileOpeningThreads())
		})
	}
}

func TestOverrideCFOpts(t *testing.T) {
	defaultOpts := newDefaultOptions()

	for _, tc := range []struct {
		desc            string
		mockAppOptions  *mockAppOptions
		writeBufferSize uint64
		numLevels       int
	}{
		{
			desc:            "override nothing",
			mockAppOptions:  newMockAppOptions(map[string]interface{}{}),
			writeBufferSize: defaultOpts.GetWriteBufferSize(),
			numLevels:       defaultOpts.GetNumLevels(),
		},
		{
			desc: "override write-buffer-size",
			mockAppOptions: newMockAppOptions(map[string]interface{}{
				writeBufferSizeCFOptName: 999_999,
			}),
			writeBufferSize: 999_999,
			numLevels:       defaultOpts.GetNumLevels(),
		},
		{
			desc: "override num-levels",
			mockAppOptions: newMockAppOptions(map[string]interface{}{
				numLevelsCFOptName: 9,
			}),
			writeBufferSize: defaultOpts.GetWriteBufferSize(),
			numLevels:       9,
		},
		{
			desc: "override write-buffer-size and num-levels",
			mockAppOptions: newMockAppOptions(map[string]interface{}{
				writeBufferSizeCFOptName: 999_999,
				numLevelsCFOptName:       9,
			}),
			writeBufferSize: 999_999,
			numLevels:       9,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			cfOpts := newDefaultOptions()
			cfOpts = overrideCFOpts(cfOpts, tc.mockAppOptions)

			require.Equal(t, tc.writeBufferSize, cfOpts.GetWriteBufferSize())
			require.Equal(t, tc.numLevels, cfOpts.GetNumLevels())
		})
	}
}

func TestReadOptsFromAppOpts(t *testing.T) {
	for _, tc := range []struct {
		desc           string
		mockAppOptions *mockAppOptions
		asyncIO        bool
	}{
		{
			desc:           "default options",
			mockAppOptions: newMockAppOptions(map[string]interface{}{}),
			asyncIO:        false,
		},
		{
			desc: "set asyncIO option to true",
			mockAppOptions: newMockAppOptions(map[string]interface{}{
				asyncIOReadOptName: true,
			}),
			asyncIO: true,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			readOpts := readOptsFromAppOpts(tc.mockAppOptions)

			require.Equal(t, tc.asyncIO, readOpts.IsAsyncIO())
		})
	}
}

func TestNewRocksDBWithOptions(t *testing.T) {
	defaultOpts := newDefaultOptions()

	name := "application"
	dir, err := os.MkdirTemp("", "rocksdb")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(dir)
		require.NoError(t, err)
	}()

	dbOpts := newDefaultOptions()
	dbOpts.SetMaxOpenFiles(999)
	cfOpts := newDefaultOptions()
	cfOpts.SetWriteBufferSize(999_999)

	db, err := newRocksDBWithOptions(name, dir, dbOpts, cfOpts, grocksdb.NewDefaultReadOptions(), true, defaultReportMetricsIntervalSecs)
	require.NoError(t, err)
	require.NoError(t, db.Close())

	dbOpts, cfOpts, err = LoadLatestOptions(filepath.Join(dir, "application.db"))
	require.NoError(t, err)
	require.Equal(t, 999, dbOpts.GetMaxOpenFiles())
	require.Equal(t, defaultOpts.GetMaxFileOpeningThreads(), dbOpts.GetMaxFileOpeningThreads())
	require.Equal(t, uint64(999_999), cfOpts.GetWriteBufferSize())
	require.Equal(t, defaultOpts.GetNumLevels(), dbOpts.GetNumLevels())
}

func TestNewDefaultOptions(t *testing.T) {
	defaultOpts := newDefaultOptions()

	maxOpenFiles := defaultOpts.GetMaxOpenFiles()
	require.Equal(t, 4096, maxOpenFiles)
}
