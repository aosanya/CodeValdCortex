package agency_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aosanya/CodeValdCortex/internal/agency"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockIntroductionRepository is a mock implementation of IntroductionRepository
type MockIntroductionRepository struct {
	mock.Mock
}

func (m *MockIntroductionRepository) GetByAgencyID(ctx context.Context, agencyID string) (*agency.AgencyIntroduction, error) {
	args := m.Called(ctx, agencyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*agency.AgencyIntroduction), args.Error(1)
}

func (m *MockIntroductionRepository) Create(ctx context.Context, intro *agency.AgencyIntroduction) error {
	args := m.Called(ctx, intro)
	return args.Error(0)
}

func (m *MockIntroductionRepository) Update(ctx context.Context, intro *agency.AgencyIntroduction) error {
	args := m.Called(ctx, intro)
	return args.Error(0)
}

func (m *MockIntroductionRepository) Delete(ctx context.Context, agencyID string) error {
	args := m.Called(ctx, agencyID)
	return args.Error(0)
}

func (m *MockIntroductionRepository) AddSection(ctx context.Context, agencyID string, section *agency.IntroductionSection) error {
	args := m.Called(ctx, agencyID, section)
	return args.Error(0)
}

func (m *MockIntroductionRepository) UpdateSection(ctx context.Context, agencyID string, section *agency.IntroductionSection) error {
	args := m.Called(ctx, agencyID, section)
	return args.Error(0)
}

func (m *MockIntroductionRepository) DeleteSection(ctx context.Context, agencyID, sectionID string) error {
	args := m.Called(ctx, agencyID, sectionID)
	return args.Error(0)
}

func (m *MockIntroductionRepository) ReorderSections(ctx context.Context, agencyID string, sectionIDs []string) error {
	args := m.Called(ctx, agencyID, sectionIDs)
	return args.Error(0)
}

// MockFlexibleIntroductionBuilder is a mock AI builder
type MockFlexibleIntroductionBuilder struct {
	mock.Mock
}

func (m *MockFlexibleIntroductionBuilder) GenerateIntroduction(ctx context.Context, template, agencyName string, keywords []string) (*agency.AgencyIntroduction, error) {
	args := m.Called(ctx, template, agencyName, keywords)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*agency.AgencyIntroduction), args.Error(1)
}

func (m *MockFlexibleIntroductionBuilder) RefineSection(ctx context.Context, section *agency.IntroductionSection, refinementInstructions string, agencyContext map[string]string) (*agency.IntroductionSection, error) {
	args := m.Called(ctx, section, refinementInstructions, agencyContext)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*agency.IntroductionSection), args.Error(1)
}

func TestIntroductionService_GetIntroduction(t *testing.T) {
	tests := []struct {
		name      string
		agencyID  string
		mockSetup func(*MockIntroductionRepository)
		wantError bool
	}{
		{
			name:     "successful retrieval",
			agencyID: "agency/123",
			mockSetup: func(m *MockIntroductionRepository) {
				m.On("GetByAgencyID", mock.Anything, "agency/123").Return(&agency.AgencyIntroduction{
					ID:       "intro/456",
					AgencyID: "agency/123",
					Template: "genesis",
				}, nil)
			},
			wantError: false,
		},
		{
			name:     "introduction not found",
			agencyID: "agency/999",
			mockSetup: func(m *MockIntroductionRepository) {
				m.On("GetByAgencyID", mock.Anything, "agency/999").Return(nil, errors.New("not found"))
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockIntroductionRepository)
			tt.mockSetup(mockRepo)

			service := agency.NewIntroductionService(mockRepo, nil)
			intro, err := service.GetIntroduction(context.Background(), tt.agencyID)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, intro)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, intro)
				assert.Equal(t, tt.agencyID, intro.AgencyID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestIntroductionService_ApplyTemplate(t *testing.T) {
	tests := []struct {
		name         string
		agencyID     string
		templateName string
		mockSetup    func(*MockIntroductionRepository)
		wantError    bool
	}{
		{
			name:         "apply genesis template successfully",
			agencyID:     "agency/123",
			templateName: "genesis",
			mockSetup: func(m *MockIntroductionRepository) {
				m.On("GetByAgencyID", mock.Anything, "agency/123").Return(nil, errors.New("not found"))
				m.On("Create", mock.Anything, mock.AnythingOfType("*agency.AgencyIntroduction")).Return(nil)
			},
			wantError: false,
		},
		{
			name:         "apply template to existing introduction",
			agencyID:     "agency/123",
			templateName: "minimal",
			mockSetup: func(m *MockIntroductionRepository) {
				m.On("GetByAgencyID", mock.Anything, "agency/123").Return(&agency.AgencyIntroduction{
					ID:       "intro/456",
					AgencyID: "agency/123",
					Template: "genesis",
				}, nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*agency.AgencyIntroduction")).Return(nil)
			},
			wantError: false,
		},
		{
			name:         "unknown template",
			agencyID:     "agency/123",
			templateName: "unknown",
			mockSetup:    func(m *MockIntroductionRepository) {},
			wantError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockIntroductionRepository)
			tt.mockSetup(mockRepo)

			service := agency.NewIntroductionService(mockRepo, nil)
			intro, err := service.ApplyTemplate(context.Background(), tt.agencyID, tt.templateName)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, intro)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, intro)
				assert.Equal(t, tt.templateName, intro.Template)
				assert.Equal(t, tt.agencyID, intro.AgencyID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestIntroductionService_AddSection(t *testing.T) {
	section := &agency.IntroductionSection{
		Type:    agency.SectionTypeText,
		Title:   "Test Section",
		Content: map[string]interface{}{"text": "Test content"},
		Order:   1,
	}

	tests := []struct {
		name      string
		agencyID  string
		section   *agency.IntroductionSection
		mockSetup func(*MockIntroductionRepository)
		wantError bool
	}{
		{
			name:     "add section successfully",
			agencyID: "agency/123",
			section:  section,
			mockSetup: func(m *MockIntroductionRepository) {
				m.On("AddSection", mock.Anything, "agency/123", section).Return(nil)
			},
			wantError: false,
		},
		{
			name:     "repository error",
			agencyID: "agency/123",
			section:  section,
			mockSetup: func(m *MockIntroductionRepository) {
				m.On("AddSection", mock.Anything, "agency/123", section).Return(errors.New("database error"))
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockIntroductionRepository)
			tt.mockSetup(mockRepo)

			service := agency.NewIntroductionService(mockRepo, nil)
			err := service.AddSection(context.Background(), tt.agencyID, tt.section)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestIntroductionService_GenerateIntroduction(t *testing.T) {
	tests := []struct {
		name      string
		agencyID  string
		template  string
		keywords  []string
		mockSetup func(*MockIntroductionRepository, *MockFlexibleIntroductionBuilder)
		wantError bool
	}{
		{
			name:     "generate introduction successfully",
			agencyID: "agency/123",
			template: "genesis",
			keywords: []string{"innovation", "technology"},
			mockSetup: func(repo *MockIntroductionRepository, builder *MockFlexibleIntroductionBuilder) {
				generatedIntro := &agency.AgencyIntroduction{
					AgencyID: "agency/123",
					Template: "genesis",
					Sections: []agency.IntroductionSection{
						{Type: agency.SectionTypeText, Title: "Overview", Content: map[string]interface{}{"text": "Generated content"}},
					},
				}
				builder.On("GenerateIntroduction", mock.Anything, "genesis", "Test Agency", []string{"innovation", "technology"}).Return(generatedIntro, nil)
				repo.On("GetByAgencyID", mock.Anything, "agency/123").Return(nil, errors.New("not found"))
				repo.On("Create", mock.Anything, mock.AnythingOfType("*agency.AgencyIntroduction")).Return(nil)
			},
			wantError: false,
		},
		{
			name:     "AI generation fails",
			agencyID: "agency/123",
			template: "genesis",
			keywords: []string{"test"},
			mockSetup: func(repo *MockIntroductionRepository, builder *MockFlexibleIntroductionBuilder) {
				builder.On("GenerateIntroduction", mock.Anything, "genesis", "Test Agency", []string{"test"}).Return(nil, errors.New("AI error"))
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockIntroductionRepository)
			mockBuilder := new(MockFlexibleIntroductionBuilder)
			tt.mockSetup(mockRepo, mockBuilder)

			service := agency.NewIntroductionService(mockRepo, mockBuilder)
			intro, err := service.GenerateIntroduction(context.Background(), tt.agencyID, tt.template, "Test Agency", tt.keywords)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, intro)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, intro)
			}

			mockRepo.AssertExpectations(t)
			mockBuilder.AssertExpectations(t)
		})
	}
}

func TestIntroductionService_RefineSection(t *testing.T) {
	section := &agency.IntroductionSection{
		ID:      "section/1",
		Type:    agency.SectionTypeText,
		Title:   "Overview",
		Content: map[string]interface{}{"text": "Original content"},
	}

	tests := []struct {
		name         string
		agencyID     string
		sectionID    string
		instructions string
		mockSetup    func(*MockIntroductionRepository, *MockFlexibleIntroductionBuilder)
		wantError    bool
	}{
		{
			name:         "refine section successfully",
			agencyID:     "agency/123",
			sectionID:    "section/1",
			instructions: "Make it more professional",
			mockSetup: func(repo *MockIntroductionRepository, builder *MockFlexibleIntroductionBuilder) {
				intro := &agency.AgencyIntroduction{
					ID:       "intro/456",
					AgencyID: "agency/123",
					Sections: []agency.IntroductionSection{*section},
				}
				refinedSection := &agency.IntroductionSection{
					ID:      "section/1",
					Type:    agency.SectionTypeText,
					Title:   "Overview",
					Content: map[string]interface{}{"text": "Refined professional content"},
				}
				repo.On("GetByAgencyID", mock.Anything, "agency/123").Return(intro, nil)
				builder.On("RefineSection", mock.Anything, section, "Make it more professional", mock.Anything).Return(refinedSection, nil)
				repo.On("UpdateSection", mock.Anything, "agency/123", refinedSection).Return(nil)
			},
			wantError: false,
		},
		{
			name:         "section not found",
			agencyID:     "agency/123",
			sectionID:    "section/999",
			instructions: "Refine this",
			mockSetup: func(repo *MockIntroductionRepository, builder *MockFlexibleIntroductionBuilder) {
				intro := &agency.AgencyIntroduction{
					ID:       "intro/456",
					AgencyID: "agency/123",
					Sections: []agency.IntroductionSection{*section},
				}
				repo.On("GetByAgencyID", mock.Anything, "agency/123").Return(intro, nil)
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockIntroductionRepository)
			mockBuilder := new(MockFlexibleIntroductionBuilder)
			tt.mockSetup(mockRepo, mockBuilder)

			service := agency.NewIntroductionService(mockRepo, mockBuilder)
			refined, err := service.RefineSection(context.Background(), tt.agencyID, tt.sectionID, tt.instructions, map[string]string{"name": "Test Agency"})

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, refined)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, refined)
			}

			mockRepo.AssertExpectations(t)
			mockBuilder.AssertExpectations(t)
		})
	}
}
