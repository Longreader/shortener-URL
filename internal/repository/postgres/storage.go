package postgres

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/Longreader/go-shortener-url.git/internal/repository"
	"github.com/Longreader/go-shortener-url.git/internal/tools"
	"github.com/sirupsen/logrus"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type PsqlStorage struct {
	db       *sqlx.DB
	deleteCh chan repository.LinkData
	wg       *sync.WaitGroup
	stop     chan struct{}
}

const (
	delBufferSize    = 50
	delBufferTimeout = time.Second
)

func NewPsqlStorage(dsn string) (*PsqlStorage, error) {

	var err error

	st := &PsqlStorage{
		deleteCh: make(chan repository.LinkData),
		stop:     make(chan struct{}),
		wg:       &sync.WaitGroup{},
	}

	st.db, err = sqlx.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = st.db.Ping()
	if err != nil {
		return nil, err
	}

	st.wg.Add(1)
	st.RunDelete()

	st.Setup()

	return st, nil
}

func (st *PsqlStorage) Setup() {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	st.db.MustExecContext(
		ctx,
		`CREATE TABLE IF NOT EXISTS links (
			id      varchar(255) NOT NULL UNIQUE,
			url     varchar(255) NOT NULL UNIQUE,
			deleted bool 		 NOT NULL DEFAULT FALSE,
			user_id uuid         NOT NULL
		);`,
	)

}

// Set method for PsqlStorage storage
func (st *PsqlStorage) Set(
	ctx context.Context,
	url repository.URL,
	user repository.User,
) (id repository.ID, err error) {

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	for {
		id, err = tools.RandStringBytes(5)
		if err != nil {
			return "", err
		}
		_, err := st.db.ExecContext(
			ctx,
			`INSERT INTO links (id, url, user_id) VALUES ($1, $2, $3) ON CONFLICT (id) DO NOTHING`,
			id, url, user,
		)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			} else {
				row := st.db.QueryRowContext(
					ctx,
					`SELECT id FROM links WHERE url=$1`,
					url,
				)
				err := row.Scan(&id)
				if err != nil {
					logrus.Printf("Error scan value %s", err)
					return "", err
				}
				return id, repository.ErrURLAlreadyExists
			}
		} else {
			break
		}
	}
	return id, nil
}

// Get method for PsqlStorage storage
func (st *PsqlStorage) Get(
	ctx context.Context,
	id repository.ID,
) (url repository.URL, deleted bool, err error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	row := st.db.QueryRowContext(
		ctx,
		`SELECT url, deleted FROM links WHERE id=$1`,
		id,
	)

	err = row.Scan(&url, &deleted)

	if err == sql.ErrNoRows {
		return "", false, repository.ErrURLNotFound
	} else if err != nil {
		return "", false, err
	}
	return url, deleted, nil
}

func (st *PsqlStorage) Delete(
	ctx context.Context,
	ids []repository.ID,
	user repository.User,
) error {
	for _, id := range ids {
		st.deleteCh <- repository.LinkData{ID: id, User: user}
	}
	return nil
}

func (st *PsqlStorage) DeleteLink(ids []repository.ID, users []repository.User) {

	ctxLocal, cancelLocal := context.WithTimeout(context.Background(), delBufferTimeout)

	_, err := st.db.ExecContext(
		ctxLocal,
		`UPDATE links SET deleted = TRUE
			FROM (SELECT UNNEST($1::text[]) AS id, UNNEST($2::uuid[]) AS user) AS data_table
			WHERE links.id = data_table.id AND user_id = data_table.user`,
		ids, users,
	)
	if err != nil {
		log.Printf("update failed: %v", err)
	}
	cancelLocal()

}

func (st *PsqlStorage) RunDelete() {

	go func() {
		ids := make([]repository.ID, 0, delBufferSize)
		users := make([]repository.User, 0, delBufferSize)
		theTicker := time.NewTicker(delBufferTimeout)
		defer theTicker.Stop()
		for {
			//timer := time.NewTimer(delBufferTimeout)
			select {
			case <-theTicker.C:
				if len(ids) != 0 || len(users) != 0 {
					st.DeleteLink(ids, users)
				}
				ids = ids[:0]
				users = users[:0]
			case <-st.stop:
				if len(ids) != 0 || len(users) != 0 {
					st.DeleteLink(ids, users)
				}
				st.wg.Done()
				return
			case data, ok := <-st.deleteCh:
				if !ok {
					st.wg.Done()
					return
				}
				if len(ids) != 0 || len(users) != 0 {
					st.DeleteLink(ids, users)
				}
				if len(ids) == delBufferSize-1 || len(users) == delBufferSize-1 {
					ids = ids[:0]
					users = users[:0]
				}
				ids = append(ids, data.ID)
				users = append(users, data.User)
			}
		}
	}()

}

func (st *PsqlStorage) GetAllByUser(
	ctx context.Context,
	user repository.User,
) (data []repository.LinkData, err error) {

	data = make([]repository.LinkData, 0)

	rows, err := st.db.QueryContext(
		ctx,
		`SELECT url, id, user_id FROM links WHERE user_id=$1 and deleted=FALSE`,
		user,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var ld repository.LinkData

		err := rows.Scan(&ld.URL, &ld.ID, &ld.User)
		if err != nil {
			return data, err
		}

		data = append(data, ld)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (st *PsqlStorage) Ping(ctx context.Context) (bool, error) {

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	err := st.db.PingContext(ctx)
	if err != nil {
		cancel()
		return false, err
	}
	return true, nil
}

func (st *PsqlStorage) Close(_ context.Context) error {

	st.stop <- struct{}{}
	st.wg.Wait()
	close(st.stop)
	return st.db.Close()
}
