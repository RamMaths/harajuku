package port

import (
  "context"
)

type FileRepository interface {
  Save(ctx context.Context, data []byte, name string) (string, error)
  Get(ctx context.Context, path string) ([]byte, error) 
  Delete(ctx context.Context, path string) error
}
