package dbaccess

import (
	"FileEngine/interfaces"
	"context"
	"database/sql"
	"time"
)

type file struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	BucketID    string     `json:"bucket_id"`
	Icon        string     `json:"icon"`
	Size        int64      `json:"size"`
	ContentType string     `json:"content_type"`
	CreateTime  *time.Time `json:"create_time"`
	UpdateTime  *time.Time `json:"update_time"`
}

type DBFile struct {
	db *sql.DB
}

func NewDBFile() interfaces.DBFile {
	return &DBFile{
		db: dbPool,
	}
}

func (d *DBFile) CreateFile(ctx context.Context, file *interfaces.FileInfo) error {
	query := `
		INSERT INTO t_file 
		(id, name, content_type, bucket_id, size, icon)
		VALUES 
		(?, ?, ?, ?, ?, ?)
	`

	_, err := d.db.ExecContext(ctx, query,
		file.ID, file.Name, file.ContentType, file.BucketID, file.Size, file.Icon)

	return err
}

func (d *DBFile) GetFileByID(ctx context.Context, fileID string) (*interfaces.FileInfo, error) {
	query := `
		SELECT 
			id, 
			name, 
			content_type, 
			bucket_id, 
			size, 
			icon, 
			create_time, 
			update_time
		FROM t_file WHERE id = ?
	`

	var file file
	err := d.db.QueryRowContext(ctx, query, fileID).Scan(
		&file.ID,
		&file.Name,
		&file.ContentType,
		&file.BucketID,
		&file.Size,
		&file.Icon,
		&file.CreateTime,
		&file.UpdateTime)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return convertToFileInfo(&file), nil
}

func (d *DBFile) GetFileByName(ctx context.Context, name string) (*interfaces.FileInfo, error) {
	query := `
		SELECT 
			id, 
			name, 
			content_type, 
			bucket_id, 
			size, 
			icon, 
			create_time, 
			update_time
		FROM t_file WHERE name = ?
	`

	var file file
	err := d.db.QueryRowContext(ctx, query, name).Scan(
		&file.ID,
		&file.Name,
		&file.ContentType,
		&file.BucketID,
		&file.Size,
		&file.Icon,
		&file.CreateTime,
		&file.UpdateTime)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return convertToFileInfo(&file), nil
}

func (d *DBFile) DeleteFile(ctx context.Context, fileID string) error {
	query := `DELETE FROM t_file WHERE id = ?`
	_, err := d.db.ExecContext(ctx, query, fileID)
	return err
}

func (d *DBFile) GetFileList(ctx context.Context, bucketID string, page, pageSize int) ([]*interfaces.FileInfo, int64, error) {
	// 获取总数
	countQuery := `SELECT COUNT(*) FROM t_file WHERE bucket_id = ?`
	var total int64
	err := d.db.QueryRowContext(ctx, countQuery, bucketID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	query := `
		SELECT 
			id, 
			name, 
			content_type, 
			bucket_id, 
			size, 
			icon, 
			create_time, 
			update_time
		FROM t_file 
		WHERE bucket_id = ?
		ORDER BY create_time DESC
		LIMIT ? OFFSET ?
	`

	rows, err := d.db.QueryContext(ctx, query, bucketID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var files []*interfaces.FileInfo
	for rows.Next() {
		var file file
		err := rows.Scan(
			&file.ID,
			&file.Name,
			&file.ContentType,
			&file.BucketID,
			&file.Size,
			&file.Icon,
			&file.CreateTime,
			&file.UpdateTime)
		if err != nil {
			return nil, 0, err
		}
		files = append(files, convertToFileInfo(&file))
	}

	return files, total, nil
}

func convertToFileInfo(file *file) *interfaces.FileInfo {
	return &interfaces.FileInfo{
		ID:          file.ID,
		Name:        file.Name,
		BucketID:    file.BucketID,
		Icon:        file.Icon,
		Size:        file.Size,
		ContentType: file.ContentType,
		CreateTime:  file.CreateTime,
		UpdateTime:  file.UpdateTime,
	}
}
