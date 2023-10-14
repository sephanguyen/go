package eureka

import (
	"context"
	"fmt"
	"sync"

	"github.com/manabie-com/backend/internal/eureka/configurations"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

func runFixBooksCurrentChapterDisplayOrder(ctx context.Context, db *pgxpool.Pool) error {
	if err := database.ExecInTx(ctx, db, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := db.Exec(ctx, `
    UPDATE  books SET
      current_chapter_display_order = fb.chapters_count
    FROM
      (
        SELECT
          b.book_id,
          chapters_count
        FROM
          (
            SELECT
              book_id,
              count(chapter_id) AS chapters_count
            FROM
              books_chapters
            WHERE
              deleted_at IS NULL
            GROUP BY
              book_id
          ) AS bcc
          JOIN books b ON b.book_id = bcc.book_id
          AND b.current_chapter_display_order != bcc.chapters_count
      ) AS fb
    WHERE
      books.book_id = fb.book_id;
    `); err != nil {
			return fmt.Errorf("db.Exec: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("database.ExecInTx: %w", err)
	}
	return nil
}

func runFixChaptersCurrentTopicDisplayOrder(ctx context.Context, db *pgxpool.Pool) error {
	if err := database.ExecInTx(ctx, db, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := db.Exec(ctx, `
    UPDATE
      chapters
    SET
      current_topic_display_order = fc.topics_count
    FROM
      (
        SELECT
          c.chapter_id,
          topics_count
        FROM
          (
            SELECT
              chapter_id,
              count(topic_id) AS topics_count
            FROM
              topics
            WHERE
              deleted_at IS NULL
            GROUP BY
              chapter_id
          ) AS cc
        JOIN chapters c ON
          c.chapter_id = cc.chapter_id
          AND c.current_topic_display_order != topics_count
      ) AS fc
    WHERE
      chapters.chapter_id = fc.chapter_id;
    `); err != nil {
			return fmt.Errorf("db.Exec: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("database.ExecInTx: %w", err)
	}
	return nil
}

func runFixTopicsLoDisplayOrderCounter(ctx context.Context, db *pgxpool.Pool) error {
	if err := database.ExecInTx(ctx, db, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := db.Exec(ctx, `
    UPDATE
      topics
    SET
      lo_display_order_counter = ft.los_count
    FROM
      (
        SELECT
          t.topic_id,
          los_count
        FROM
          (
            SELECT
              topic_id,
              count(lo_id) AS los_count
            FROM
              topics_learning_objectives
            WHERE
              deleted_at IS NULL
            GROUP BY
              topic_id
          ) AS tloc
        JOIN topics t ON
          t.topic_id = tloc.topic_id
          AND t.lo_display_order_counter != tloc.los_count
      ) AS ft
    WHERE
      topics.topic_id = ft.topic_id;
    `); err != nil {
			return fmt.Errorf("db.Exec: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("database.ExecInTx: %w", err)
	}
	return nil
}

func runFixChaptersDisplayOrder(ctx context.Context, db *pgxpool.Pool) error {
	if err := database.ExecInTx(ctx, db, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := db.Exec(ctx, `
    UPDATE
      chapters
    SET
      display_order = fc.new_display_order
    FROM
      (
        SELECT
          book_id,
          chapter_id,
          new_display_order
        FROM
          (
            SELECT
              b.book_id,
              bc.updated_at,
              c.chapter_id,
              c.display_order,
              ROW_NUMBER() OVER (PARTITION BY bc.book_id ORDER BY c.display_order) AS new_display_order
            FROM
              chapters c
            INNER JOIN books_chapters bc ON bc.chapter_id = c.chapter_id
            INNER JOIN books b ON b.book_id = bc.book_id
            WHERE
              COALESCE(
                c.deleted_at,
                bc.deleted_at,
                b.deleted_at
              ) IS NULL
            ORDER BY
              bc.updated_at
          ) AS bcd
        WHERE
          display_order != new_display_order
      ) AS fc
    WHERE
      chapters.chapter_id = fc.chapter_id;
    `); err != nil {
			return fmt.Errorf("db.Exec: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("database.ExecInTx: %w", err)
	}
	return nil
}

func runFixTopicsDisplayOrder(ctx context.Context, db *pgxpool.Pool) error {
	if err := database.ExecInTx(ctx, db, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := db.Exec(ctx, `
    UPDATE
      topics
    SET
      display_order = ft.new_display_order
    FROM
      (
        SELECT
          td.topic_id,
          td.new_display_order
        FROM
          (
            SELECT
              topic_id,
              display_order,
              chapter_id,
              ROW_NUMBER() OVER (PARTITION BY chapter_id ORDER BY display_order) AS new_display_order
            FROM
              topics
            WHERE
              chapter_id IS NOT NULL
              AND deleted_at IS NULL
          ) AS td
        WHERE
          td.display_order != td.new_display_order
      ) AS ft
    WHERE
      topics.topic_id = ft.topic_id;
    `); err != nil {
			return fmt.Errorf("db.Exec: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("database.ExecInTx: %w", err)
	}
	return nil
}

func runFixTopicsLODisplayOrder(ctx context.Context, db *pgxpool.Pool) error {
	if err := database.ExecInTx(ctx, db, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := db.Exec(ctx, `
    UPDATE
      topics_learning_objectives tlo
    SET
      display_order = ftlo.new_display_order
    FROM
      (
        SELECT
          topic_id,
          lo_id,
          new_display_order
        FROM
          (
            SELECT
              topic_id,
              lo_id,
              display_order,
              ROW_NUMBER() OVER(PARTITION BY topic_id ORDER BY display_order) AS new_display_order
            FROM
              topics_learning_objectives
            WHERE
              deleted_at IS NULL
          ) AS tlod
        WHERE
          display_order != new_display_order
      ) AS ftlo
    WHERE
      tlo.topic_id = ftlo.topic_id
      AND tlo.lo_id = ftlo.lo_id;
    `); err != nil {
			return fmt.Errorf("db.Exec: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("database.ExecInTx: %w", err)
	}
	return nil
}

func init() {
	bootstrap.RegisterJob("eureka_fix_current_display_order", RunFixCurrentDisplayOrder)
}

func RunFixCurrentDisplayOrder(ctx context.Context, _ configurations.Config, rsc *bootstrap.Resources) error {
	zapLogger := rsc.Logger()

	eurekaDB := rsc.DB().DB.(*pgxpool.Pool)

	var wg sync.WaitGroup
	wg.Add(6)
	go func() {
		defer wg.Done()
		if err := runFixBooksCurrentChapterDisplayOrder(ctx, eurekaDB); err != nil {
			zapLogger.Error("FixCurrentDisplayOrder: runFixBooksCurrentChapterDisplayOrder", zap.Error(err))
		}
	}()
	go func() {
		defer wg.Done()
		if err := runFixChaptersCurrentTopicDisplayOrder(ctx, eurekaDB); err != nil {
			zapLogger.Error("FixCurrentDisplayOrder: runFixChaptersCurrentTopicDisplayOrder", zap.Error(err))
		}
	}()
	go func() {
		defer wg.Done()
		if err := runFixTopicsLoDisplayOrderCounter(ctx, eurekaDB); err != nil {
			zapLogger.Error("FixCurrentDisplayOrder: runFixChaptersCurrentTopicDisplayOrder", zap.Error(err))
		}
	}()
	go func() {
		defer wg.Done()
		if err := runFixChaptersDisplayOrder(ctx, eurekaDB); err != nil {
			zapLogger.Error("FixCurrentDisplayOrder: runFixChaptersDisplayOrder", zap.Error(err))
		}
	}()
	go func() {
		defer wg.Done()
		if err := runFixTopicsDisplayOrder(ctx, eurekaDB); err != nil {
			zapLogger.Error("FixCurrentDisplayOrder: runFixTopicsDisplayOrder", zap.Error(err))
		}
	}()
	go func() {
		defer wg.Done()
		if err := runFixTopicsLODisplayOrder(ctx, eurekaDB); err != nil {
			zapLogger.Error("FixCurrentDisplayOrder: runFixTopicsLODisplayOrder", zap.Error(err))
		}
	}()
	wg.Wait()
	return nil
}
