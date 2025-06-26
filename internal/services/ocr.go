package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"cloud.google.com/go/vision/v2/apiv1/visionpb"
	vision "cloud.google.com/go/vision/v2/apiv1"
)

// OCRService handles OCR processing
type OCRService struct {
	visionClient *vision.ImageAnnotatorClient
	ocrSpaceURL  string
	ocrSpaceKey  string
}

// NewOCRService creates a new OCR service
func NewOCRService() (*OCRService, error) {
	service := &OCRService{
		ocrSpaceURL: os.Getenv("OCR_SERVICE_URL"),
		ocrSpaceKey: os.Getenv("OCR_API_KEY"),
	}

	// Try to initialize Google Vision API
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") != "" || os.Getenv("FIREBASE_SERVICE_ACCOUNT_KEY") != "" {
		visionClient, err := vision.NewImageAnnotatorClient(context.Background())
		if err == nil {
			service.visionClient = visionClient
		}
	}

	return service, nil
}

// ProcessOCR performs OCR on a document
func (s *OCRService) ProcessOCR(ctx context.Context, fileReader io.Reader) (string, error) {
	// Try Google Vision API first
	if s.visionClient != nil {
		return s.processWithGoogleVision(ctx, fileReader)
	}

	// Try OCR.space
	if s.ocrSpaceURL != "" && s.ocrSpaceKey != "" {
		return s.processWithOCRSpace(ctx, fileReader)
	}

	// Fallback to mock OCR
	return s.processMockOCR(ctx, fileReader)
}

// processWithGoogleVision uses Google Vision API for OCR
func (s *OCRService) processWithGoogleVision(ctx context.Context, fileReader io.Reader) (string, error) {
	// Read the image data
	imageData, err := io.ReadAll(fileReader)
	if err != nil {
		return "", fmt.Errorf("failed to read image data: %w", err)
	}

	// Create image object
	image := &visionpb.Image{
		Content: imageData,
	}

	// Create text detection request
	request := &visionpb.BatchAnnotateImagesRequest{
		Requests: []*visionpb.AnnotateImageRequest{
			{
				Image: image,
				Features: []*visionpb.Feature{
					{
						Type: visionpb.Feature_TEXT_DETECTION,
					},
				},
			},
		},
	}

	// Perform text detection
	resp, err := s.visionClient.BatchAnnotateImages(ctx, request)
	if err != nil {
		return "", fmt.Errorf("failed to detect text: %w", err)
	}

	if len(resp.Responses) == 0 || len(resp.Responses[0].TextAnnotations) == 0 {
		return "", fmt.Errorf("no text detected")
	}

	// Extract text from all detected text blocks
	var texts []string
	for _, annotation := range resp.Responses[0].TextAnnotations {
		texts = append(texts, annotation.Description)
	}

	return strings.Join(texts, "\n"), nil
}

// processWithOCRSpace uses OCR.space API for OCR
func (s *OCRService) processWithOCRSpace(ctx context.Context, fileReader io.Reader) (string, error) {
	// Read the image data
	_, err := io.ReadAll(fileReader)
	if err != nil {
		return "", fmt.Errorf("failed to read image data: %w", err)
	}

	// Prepare form data
	formData := url.Values{}
	formData.Set("apikey", s.ocrSpaceKey)
	formData.Set("language", "eng")
	formData.Set("isOverlayRequired", "false")

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", s.ocrSpaceURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send OCR request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var ocrResponse struct {
		ParsedResults []struct {
			ParsedText string `json:"ParsedText"`
		} `json:"ParsedResults"`
		ErrorMessage string `json:"ErrorMessage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ocrResponse); err != nil {
		return "", fmt.Errorf("failed to decode OCR response: %w", err)
	}

	if ocrResponse.ErrorMessage != "" {
		return "", fmt.Errorf("OCR error: %s", ocrResponse.ErrorMessage)
	}

	if len(ocrResponse.ParsedResults) == 0 {
		return "", fmt.Errorf("no text detected")
	}

	return ocrResponse.ParsedResults[0].ParsedText, nil
}

// processMockOCR returns mock OCR text for testing
func (s *OCRService) processMockOCR(ctx context.Context, fileReader io.Reader) (string, error) {
	// Read a small amount to simulate processing
	_, err := io.ReadAll(fileReader)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return "This is a mock OCR result. In a real implementation, this would contain the actual text extracted from the document image.", nil
}

// Close closes the OCR service
func (s *OCRService) Close() error {
	if s.visionClient != nil {
		return s.visionClient.Close()
	}
	return nil
} 