package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"catalogizer/models"
	"catalogizer/repository"
)

type ConversionService struct {
	conversionRepo *repository.ConversionRepository
	userRepo       *repository.UserRepository
	authService    *AuthService
}

func NewConversionService(conversionRepo *repository.ConversionRepository, userRepo *repository.UserRepository, authService *AuthService) *ConversionService {
	return &ConversionService{
		conversionRepo: conversionRepo,
		userRepo:       userRepo,
		authService:    authService,
	}
}

func (s *ConversionService) CreateConversionJob(userID int, request *models.ConversionRequest) (*models.ConversionJob, error) {
	if !s.validateConversionRequest(request) {
		return nil, fmt.Errorf("invalid conversion request")
	}

	job := &models.ConversionJob{
		UserID:           userID,
		SourcePath:       request.SourcePath,
		TargetPath:       request.TargetPath,
		SourceFormat:     request.SourceFormat,
		TargetFormat:     request.TargetFormat,
		ConversionType:   request.ConversionType,
		Quality:          request.Quality,
		Settings:         request.Settings,
		Priority:         request.Priority,
		Status:           models.ConversionStatusPending,
		CreatedAt:        time.Now(),
		ScheduledFor:     request.ScheduledFor,
	}

	id, err := s.conversionRepo.CreateJob(job)
	if err != nil {
		return nil, fmt.Errorf("failed to create conversion job: %w", err)
	}

	job.ID = id
	return job, nil
}

func (s *ConversionService) StartConversion(jobID int) error {
	job, err := s.conversionRepo.GetJob(jobID)
	if err != nil {
		return fmt.Errorf("failed to get conversion job: %w", err)
	}

	if job.Status != models.ConversionStatusPending {
		return fmt.Errorf("job is not in pending status")
	}

	job.Status = models.ConversionStatusRunning
	job.StartedAt = &time.Time{}
	*job.StartedAt = time.Now()

	err = s.conversionRepo.UpdateJob(job)
	if err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	go s.processConversion(job)

	return nil
}

func (s *ConversionService) processConversion(job *models.ConversionJob) {
	var err error

	defer func() {
		if r := recover(); r != nil {
			s.handleConversionError(job, fmt.Errorf("conversion panic: %v", r))
		}
	}()

	switch job.ConversionType {
	case models.ConversionTypeVideo:
		err = s.convertVideo(job)
	case models.ConversionTypeAudio:
		err = s.convertAudio(job)
	case models.ConversionTypeDocument:
		err = s.convertDocument(job)
	case models.ConversionTypeImage:
		err = s.convertImage(job)
	default:
		err = fmt.Errorf("unsupported conversion type: %s", job.ConversionType)
	}

	if err != nil {
		s.handleConversionError(job, err)
		return
	}

	s.handleConversionSuccess(job)
}

func (s *ConversionService) convertVideo(job *models.ConversionJob) error {
	args := s.buildFFmpegVideoArgs(job)

	cmd := exec.Command("ffmpeg", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("ffmpeg video conversion failed: %w", err)
	}

	return nil
}

func (s *ConversionService) convertAudio(job *models.ConversionJob) error {
	args := s.buildFFmpegAudioArgs(job)

	cmd := exec.Command("ffmpeg", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("ffmpeg audio conversion failed: %w", err)
	}

	return nil
}

func (s *ConversionService) convertDocument(job *models.ConversionJob) error {
	switch {
	case s.isEbookConversion(job):
		return s.convertEbook(job)
	case s.isPDFConversion(job):
		return s.convertPDF(job)
	default:
		return fmt.Errorf("unsupported document conversion")
	}
}

func (s *ConversionService) convertEbook(job *models.ConversionJob) error {
	args := []string{
		job.SourcePath,
		job.TargetPath,
	}

	if job.Settings != nil {
		settings := make(map[string]interface{})
		if err := json.Unmarshal([]byte(*job.Settings), &settings); err == nil {
			if cover, ok := settings["preserve_cover"].(bool); ok && cover {
				args = append(args, "--preserve-cover")
			}
			if metadata, ok := settings["preserve_metadata"].(bool); ok && metadata {
				args = append(args, "--preserve-metadata")
			}
		}
	}

	cmd := exec.Command("ebook-convert", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("ebook conversion failed: %w", err)
	}

	return nil
}

func (s *ConversionService) convertPDF(job *models.ConversionJob) error {
	// Placeholder for PDF conversion using tools like pandoc or libreoffice
	return fmt.Errorf("PDF conversion not yet implemented")
}

func (s *ConversionService) convertImage(job *models.ConversionJob) error {
	args := s.buildImageMagickArgs(job)

	cmd := exec.Command("convert", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("imagemagick conversion failed: %w", err)
	}

	return nil
}

func (s *ConversionService) buildFFmpegVideoArgs(job *models.ConversionJob) []string {
	args := []string{
		"-i", job.SourcePath,
		"-y", // Overwrite output file
	}

	switch job.Quality {
	case "low":
		args = append(args, "-crf", "28", "-preset", "fast")
	case "medium":
		args = append(args, "-crf", "23", "-preset", "medium")
	case "high":
		args = append(args, "-crf", "18", "-preset", "slow")
	case "lossless":
		args = append(args, "-crf", "0", "-preset", "veryslow")
	default:
		args = append(args, "-crf", "23", "-preset", "medium")
	}

	if job.Settings != nil {
		settings := make(map[string]interface{})
		if err := json.Unmarshal([]byte(*job.Settings), &settings); err == nil {
			if resolution, ok := settings["resolution"].(string); ok {
				args = append(args, "-s", resolution)
			}
			if framerate, ok := settings["framerate"].(string); ok {
				args = append(args, "-r", framerate)
			}
			if bitrate, ok := settings["bitrate"].(string); ok {
				args = append(args, "-b:v", bitrate)
			}
		}
	}

	args = append(args, job.TargetPath)
	return args
}

func (s *ConversionService) buildFFmpegAudioArgs(job *models.ConversionJob) []string {
	args := []string{
		"-i", job.SourcePath,
		"-y", // Overwrite output file
	}

	switch job.Quality {
	case "low":
		args = append(args, "-ab", "96k")
	case "medium":
		args = append(args, "-ab", "192k")
	case "high":
		args = append(args, "-ab", "320k")
	case "lossless":
		args = append(args, "-c:a", "flac")
	default:
		args = append(args, "-ab", "192k")
	}

	if job.Settings != nil {
		settings := make(map[string]interface{})
		if err := json.Unmarshal([]byte(*job.Settings), &settings); err == nil {
			if sampleRate, ok := settings["sample_rate"].(string); ok {
				args = append(args, "-ar", sampleRate)
			}
			if channels, ok := settings["channels"].(string); ok {
				args = append(args, "-ac", channels)
			}
		}
	}

	args = append(args, job.TargetPath)
	return args
}

func (s *ConversionService) buildImageMagickArgs(job *models.ConversionJob) []string {
	args := []string{job.SourcePath}

	if job.Settings != nil {
		settings := make(map[string]interface{})
		if err := json.Unmarshal([]byte(*job.Settings), &settings); err == nil {
			if resize, ok := settings["resize"].(string); ok {
				args = append(args, "-resize", resize)
			}
			if quality, ok := settings["quality"].(string); ok {
				args = append(args, "-quality", quality)
			}
			if compress, ok := settings["compress"].(bool); ok && compress {
				args = append(args, "-compress", "JPEG")
			}
		}
	}

	args = append(args, job.TargetPath)
	return args
}

func (s *ConversionService) handleConversionSuccess(job *models.ConversionJob) {
	job.Status = models.ConversionStatusCompleted
	job.CompletedAt = &time.Time{}
	*job.CompletedAt = time.Now()

	if job.StartedAt != nil {
		duration := job.CompletedAt.Sub(*job.StartedAt)
		job.Duration = &duration
	}

	err := s.conversionRepo.UpdateJob(job)
	if err != nil {
		fmt.Printf("Failed to update completed job %d: %v\n", job.ID, err)
	}

	s.notifyUser(job, "Conversion completed successfully")
}

func (s *ConversionService) handleConversionError(job *models.ConversionJob, conversionError error) {
	job.Status = models.ConversionStatusFailed
	job.CompletedAt = &time.Time{}
	*job.CompletedAt = time.Now()
	errorMsg := conversionError.Error()
	job.ErrorMessage = &errorMsg

	if job.StartedAt != nil {
		duration := job.CompletedAt.Sub(*job.StartedAt)
		job.Duration = &duration
	}

	err := s.conversionRepo.UpdateJob(job)
	if err != nil {
		fmt.Printf("Failed to update failed job %d: %v\n", job.ID, err)
	}

	s.notifyUser(job, fmt.Sprintf("Conversion failed: %s", conversionError.Error()))
}

func (s *ConversionService) notifyUser(job *models.ConversionJob, message string) {
	// In a full implementation, this would send notifications via email, push, etc.
	fmt.Printf("Notification for user %d: %s (Job %d)\n", job.UserID, message, job.ID)
}

func (s *ConversionService) GetUserJobs(userID int, status *string, limit, offset int) ([]models.ConversionJob, error) {
	return s.conversionRepo.GetUserJobs(userID, status, limit, offset)
}

func (s *ConversionService) GetJob(jobID int, userID int) (*models.ConversionJob, error) {
	job, err := s.conversionRepo.GetJob(jobID)
	if err != nil {
		return nil, err
	}

	if job.UserID != userID {
		hasPermission, err := s.authService.CheckPermission(userID, models.PermissionViewMedia)
		if err != nil || !hasPermission {
			return nil, fmt.Errorf("unauthorized to view this job")
		}
	}

	return job, nil
}

func (s *ConversionService) CancelJob(jobID int, userID int) error {
	job, err := s.conversionRepo.GetJob(jobID)
	if err != nil {
		return err
	}

	if job.UserID != userID {
		hasPermission, err := s.authService.CheckPermission(userID, models.PermissionManageUsers)
		if err != nil || !hasPermission {
			return fmt.Errorf("unauthorized to cancel this job")
		}
	}

	if job.Status != models.ConversionStatusPending && job.Status != models.ConversionStatusRunning {
		return fmt.Errorf("cannot cancel job in status: %s", job.Status)
	}

	job.Status = models.ConversionStatusCancelled
	job.CompletedAt = &time.Time{}
	*job.CompletedAt = time.Now()

	return s.conversionRepo.UpdateJob(job)
}

func (s *ConversionService) RetryJob(jobID int, userID int) error {
	job, err := s.conversionRepo.GetJob(jobID)
	if err != nil {
		return err
	}

	if job.UserID != userID {
		hasPermission, err := s.authService.CheckPermission(userID, models.PermissionManageUsers)
		if err != nil || !hasPermission {
			return fmt.Errorf("unauthorized to retry this job")
		}
	}

	if job.Status != models.ConversionStatusFailed {
		return fmt.Errorf("can only retry failed jobs")
	}

	job.Status = models.ConversionStatusPending
	job.StartedAt = nil
	job.CompletedAt = nil
	job.Duration = nil
	job.ErrorMessage = nil

	err = s.conversionRepo.UpdateJob(job)
	if err != nil {
		return err
	}

	return s.StartConversion(jobID)
}

func (s *ConversionService) GetJobStatistics(userID *int, startDate, endDate time.Time) (*models.ConversionStatistics, error) {
	return s.conversionRepo.GetStatistics(userID, startDate, endDate)
}

func (s *ConversionService) CleanupCompletedJobs(olderThan time.Time) error {
	return s.conversionRepo.CleanupJobs(olderThan)
}

func (s *ConversionService) GetSupportedFormats() *models.SupportedFormats {
	return &models.SupportedFormats{
		Video: models.VideoFormats{
			Input:  []string{"mp4", "avi", "mkv", "mov", "wmv", "flv", "webm", "m4v", "3gp"},
			Output: []string{"mp4", "avi", "mkv", "mov", "webm", "m4v"},
		},
		Audio: models.AudioFormats{
			Input:  []string{"mp3", "wav", "flac", "aac", "ogg", "wma", "m4a", "opus"},
			Output: []string{"mp3", "wav", "flac", "aac", "ogg", "m4a", "opus"},
		},
		Document: models.DocumentFormats{
			Input:  []string{"epub", "mobi", "azw", "azw3", "pdf", "txt", "docx", "odt"},
			Output: []string{"epub", "mobi", "pdf", "txt", "html"},
		},
		Image: models.ImageFormats{
			Input:  []string{"jpg", "jpeg", "png", "gif", "bmp", "tiff", "webp", "svg"},
			Output: []string{"jpg", "jpeg", "png", "gif", "bmp", "tiff", "webp"},
		},
	}
}

func (s *ConversionService) validateConversionRequest(request *models.ConversionRequest) bool {
	if request.SourcePath == "" || request.TargetPath == "" {
		return false
	}

	if request.SourceFormat == "" || request.TargetFormat == "" {
		return false
	}

	if request.ConversionType == "" {
		return false
	}

	if !s.isValidConversionType(request.ConversionType) {
		return false
	}

	if !s.isSupportedFormat(request.ConversionType, request.SourceFormat, request.TargetFormat) {
		return false
	}

	return true
}

func (s *ConversionService) isValidConversionType(conversionType string) bool {
	validTypes := []string{
		models.ConversionTypeVideo,
		models.ConversionTypeAudio,
		models.ConversionTypeDocument,
		models.ConversionTypeImage,
	}

	for _, validType := range validTypes {
		if conversionType == validType {
			return true
		}
	}

	return false
}

func (s *ConversionService) isSupportedFormat(conversionType, sourceFormat, targetFormat string) bool {
	formats := s.GetSupportedFormats()

	switch conversionType {
	case models.ConversionTypeVideo:
		return s.isFormatSupported(sourceFormat, formats.Video.Input) && s.isFormatSupported(targetFormat, formats.Video.Output)
	case models.ConversionTypeAudio:
		return s.isFormatSupported(sourceFormat, formats.Audio.Input) && s.isFormatSupported(targetFormat, formats.Audio.Output)
	case models.ConversionTypeDocument:
		return s.isFormatSupported(sourceFormat, formats.Document.Input) && s.isFormatSupported(targetFormat, formats.Document.Output)
	case models.ConversionTypeImage:
		return s.isFormatSupported(sourceFormat, formats.Image.Input) && s.isFormatSupported(targetFormat, formats.Image.Output)
	}

	return false
}

func (s *ConversionService) isFormatSupported(format string, supportedFormats []string) bool {
	format = strings.ToLower(format)
	for _, supported := range supportedFormats {
		if format == strings.ToLower(supported) {
			return true
		}
	}
	return false
}

func (s *ConversionService) isEbookConversion(job *models.ConversionJob) bool {
	ebookFormats := []string{"epub", "mobi", "azw", "azw3", "txt", "html"}
	return s.isFormatSupported(job.SourceFormat, ebookFormats) || s.isFormatSupported(job.TargetFormat, ebookFormats)
}

func (s *ConversionService) isPDFConversion(job *models.ConversionJob) bool {
	return job.SourceFormat == "pdf" || job.TargetFormat == "pdf"
}

func (s *ConversionService) GetJobQueue() ([]models.ConversionJob, error) {
	return s.conversionRepo.GetJobsByStatus(models.ConversionStatusPending, 100, 0)
}

func (s *ConversionService) ProcessJobQueue() error {
	jobs, err := s.GetJobQueue()
	if err != nil {
		return err
	}

	for _, job := range jobs {
		if job.ScheduledFor != nil && job.ScheduledFor.After(time.Now()) {
			continue
		}

		err := s.StartConversion(job.ID)
		if err != nil {
			fmt.Printf("Failed to start conversion job %d: %v\n", job.ID, err)
		}
	}

	return nil
}