package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func GetCertDenomNm(denomID string) string {
	switch denomID {
	case "CertificateAuditing":
		return "Auditing"
	case "CertificateIdentity":
		return "Identity"
	default:
		return ""
	}
}

// GetCertifier returns certificer of the certificate.
func (c Certificate) GetCertifier() sdk.AccAddress {
	certifierAddr, err := sdk.AccAddressFromBech32(c.Certifier)
	if err != nil {
		panic(err)
	}
	return certifierAddr
}

const (
	CertificateSchema = `
	{
		"$schema": "http://json-schema.org/draft-04/schema#",
		"title": "certificate-schema",
		"description": "Certificate NFT Schema",
		"type": "object",
		"properties": {
			"certificate_id": {
				"description": "unique certificate ID issued",
				"type": "integer",
				"minimum": "1",
			},
			"content": {
				"description": "content of certificate",
				"type": "string",
			},
			"description": {
				"description": "description of certificate",
				"type": "string",
			},
			"certifier": {
				"description": "certifier address",
				"type": "string",
			}
		},
		"additionalProperties": false,
		"required": [
			"certificate_id",
			"content",
			"description",
			"certifier"
		]
	}`
)
