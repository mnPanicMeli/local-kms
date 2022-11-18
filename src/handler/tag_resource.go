package handler

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/mnPanicMeli/local-kms/src/cmk"
	"github.com/mnPanicMeli/local-kms/src/data"
)

func (r *RequestHandler) TagResource() Response {

	var body *kms.TagResourceInput
	err := r.decodeBodyInto(&body)

	if err != nil {
		body = &kms.TagResourceInput{}
	}

	//--------------------------------
	// Validation

	if body.KeyId == nil {
		msg := "1 validation error detected: Value null at 'keyId' failed to satisfy constraint: Member must not be null"

		r.logger.Warnf(msg)
		return NewValidationExceptionResponse(msg)
	}

	if body.Tags == nil {
		msg := "1 validation error detected: Value null at 'tags' failed to satisfy constraint: Member must not be null"

		r.logger.Warnf(msg)
		return NewValidationExceptionResponse(msg)
	}

	response := r.validateTags(body.Tags)
	if !response.Empty() {
		return response
	}

	//---

	key, response := r.getKey(*body.KeyId)
	if !response.Empty() {
		return response
	}

	switch key.GetMetadata().KeyState {
	case cmk.KeyStatePendingDeletion:
		msg := fmt.Sprintf("%s is pending deletion.", *body.KeyId)

		r.logger.Warnf(msg)
		return NewKMSInvalidStateExceptionResponse(msg)

	}

	//--------------------------------
	// Create the tags

	if body.Tags != nil && len(body.Tags) > 0 {
		for _, kv := range body.Tags {
			t := &data.Tag{
				TagKey:   *kv.TagKey,
				TagValue: *kv.TagValue,
			}

			_ = r.database.SaveTag(key, t)

			r.logger.Infof("New tag created: %s / %s\n", t.TagKey, t.TagValue)
		}
	}

	//---

	return NewResponse(200, nil)
}
