package harness

import (
	"bufio"
	"crypto/sha1"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"github.intel.com/hpdd/logging/debug"
)

// FileChecksum is a calculated checksum
type FileChecksum [20]byte

// TestFile describes a test file generated by the harness
type TestFile struct {
	Path     string
	Checksum FileChecksum
}

// GetFileChecksum returns a checksum of the file's data
func GetFileChecksum(filePath string) (FileChecksum, error) {
	buf, err := ioutil.ReadFile(filePath)
	if err != nil {
		return FileChecksum{}, errors.Wrapf(err, "Couldn't get checksum for %s", filePath)
	}

	return sha1.Sum(buf), nil
}

// NewTestFile generates a new test file
func NewTestFile(dir, prefix string) (*TestFile, error) {
	// Let's try copying the contents of the test executable
	// out as the test file. More interesting than an empty
	// file or a bunch of zeros.
	out, err := ioutil.TempFile(dir, prefix)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create test file")
	}

	// Won't work on OS X, but then again none of this will...
	srcPath, err := os.Readlink("/proc/self/exe")
	if err != nil {
		return nil, errors.Wrap(err, "Unable to find path to self")
	}

	in, err := os.Open(srcPath)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to open %s for read", srcPath)
	}
	defer in.Close()

	if _, err := bufio.NewReader(in).WriteTo(out); err != nil {
		return nil, errors.Wrap(err, "Failed to write data to test file")
	}

	debug.Printf("Created test file: %s", out.Name())
	sum, err := GetFileChecksum(out.Name())
	if err != nil {
		return nil, err
	}

	return &TestFile{
		Path:     out.Name(),
		Checksum: sum,
	}, out.Close()
}

// CreateTestfile creates a test file and adds its path to the context's
// cleanup queue.
func (ctx *ScenarioContext) CreateTestfile(dir, key string) (string, error) {
	tf, err := NewTestFile(dir, key)
	if err != nil {
		return "", errors.Wrap(err, "Could not generate test file")
	}

	ctx.AddCleanup(func() error {
		return os.Remove(tf.Path)
	})
	ctx.TestFiles[key] = tf

	return tf.Path, nil
}