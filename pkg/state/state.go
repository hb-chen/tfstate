package state

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hb-chen/tfstate/pkg/sync"
	"github.com/hb-chen/tfstate/pkg/sync/memory"
	"github.com/labstack/gommon/log"
)

type State interface {
	Get(id string) ([]byte, error)
	Update(id string, data []byte) error
	Lock(id, token string) error
	Unlock(id, token string) error
}

type state struct {
	sync sync.Sync
}

func (*state) Get(id string) ([]byte, error) {
	outFile := "./data/" + id + ".tfstate"
	data, err := ioutil.ReadFile(outFile)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (*state) Update(id string, data []byte) error {
	// 输出文件夹不存在时创建
	outFile := "./data/" + id + ".tfstate"
	dir := filepath.Dir(outFile)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0760); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	tempFile, err := ioutil.TempFile(dir, filepath.Base(outFile)+".")
	if err != nil {
		return err
	}
	defer func() {
		if err = tempFile.Close(); err != nil {
			if e, ok := err.(*os.PathError); !ok || e.Err != os.ErrClosed {
				log.Info(err)
			}
		}
	}()

	_, err = tempFile.Write(data)
	if err != nil {
		return err
	}

	if err = tempFile.Chmod(0644); err != nil {
		return err
	}

	// Close the file immediately for platforms (eg. Windows) that cannot move
	// a file while a process is holding a file handle.
	err = tempFile.Close()
	if err != nil {
		return err
	}

	err = os.Rename(tempFile.Name(), outFile)
	if err != nil {
		return err
	}

	return nil
}

func (s *state) Lock(id, token string) error {
	_ = s.sync.Lock(id, sync.LockToken(token))

	return nil
}

func (s *state) Unlock(id, token string) error {
	_ = s.sync.Unlock(id, sync.UnlockToken(token))

	return nil
}

func NewState() State {
	return &state{sync: memory.NewSync()}
}
