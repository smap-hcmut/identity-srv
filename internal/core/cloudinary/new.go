package cloudinary

import (
	"fmt"

	"github.com/cloudinary/cloudinary-go/v2"
)

type implCloudinary struct {
	cld *cloudinary.Cloudinary
}

func New(cloudName string, apiKey string, apiSecret string) (Usecase, error) {
	if cloudName == "" || apiKey == "" || apiSecret == "" {
		return nil, fmt.Errorf("invalid cloudinary configuration")
	}

	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to create cloudinary client: %w", err)
	}

	return &implCloudinary{
		cld: cld,
	}, nil
}
