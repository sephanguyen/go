package usermgmt

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	spb "github.com/manabie-com/backend/pkg/manabuf/shamir/v1"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	AESKey = "yYdV82bsdXO%Cl2Uq5F^^19GUh8%^W3j"
	AESIV  = "^30F6l#gm0C!@oD7"
)

func (s *suite) aValidVerifySignatureRequest(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var publicKey *string
	body := []byte("payload")

	stmt := `SELECT public_key from api_keypair WHERE user_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC LIMIT 1`
	err := database.Select(ctx, s.BobPostgresDBTrace, stmt, database.Text(stepState.CurrentUserID)).ScanFields(&publicKey)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "database.Select")
	}
	repo := &repository.DomainAPIKeypairRepo{
		EncryptedKey:  AESKey,
		InitialVector: AESIV,
	}
	apiKeypair, err := repo.GetByPublicKey(ctx, s.BobPostgresDBTrace, *publicKey)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "repo.GetByPublicKey")
	}

	mac := hmac.New(sha256.New, []byte(apiKeypair.PrivateKey().String()))
	_, err = mac.Write(body)
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrap(err, "mac.Write").Error())
	}
	signature := hex.EncodeToString(mac.Sum(nil))

	stepState.Request = &spb.VerifySignatureRequest{
		PublicKey: *publicKey,
		Signature: signature,
		Body:      body,
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) anInvalidVerifySignatureRequestWith(ctx context.Context, condition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &spb.VerifySignatureRequest{}

	var publicKey *string
	stmt := `SELECT public_key from api_keypair WHERE user_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC LIMIT 1`
	err := database.Select(ctx, s.BobPostgresDBTrace, stmt, database.Text(stepState.CurrentUserID)).ScanFields(&publicKey)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "database.Select")
	}
	req.PublicKey = *publicKey
	switch condition {
	case "invalid signature":
		req.Signature = "invalid-signature"
	case "invalid public key":
		req.PublicKey = "invalid-public-key"
	case "api key was deleted":
		stmt := `UPDATE api_keypair SET deleted_at = now() WHERE user_id = $1 AND deleted_at IS NULL`
		_, err := s.BobPostgresDBTrace.Exec(ctx, stmt, database.Text(stepState.CurrentUserID))
		if err != nil {
			return StepStateToContext(ctx, stepState), errors.Wrap(err, "s.BobPostgresDBTrace.Exec")
		}
	}

	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aClientVerifiesSignature(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = spb.NewTokenReaderServiceClient(s.ShamirConn).VerifySignature(ctx, stepState.Request.(*spb.VerifySignatureRequest))

	return StepStateToContext(ctx, stepState), nil
}
