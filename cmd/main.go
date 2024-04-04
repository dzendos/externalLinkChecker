package main

import (
	"context"
	"externalLinkChecker/internal/config"
	"externalLinkChecker/internal/pkg/service/comment_parser"
	"externalLinkChecker/internal/pkg/service/repo_parser"
	"externalLinkChecker/internal/pkg/storage"
)

func main() {
	ctx := context.Background()

	cfg := config.Load()

	st := storage.New()
	defer st.Close()

	cp := comment_parser.New(st)

	rp := repo_parser.New(ctx, cfg, cp, st)
	rp.Run(ctx)
}
