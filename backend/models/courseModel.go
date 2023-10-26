package models

import (
	"path/filepath"
	"time"

	"github.com/mspcix/google-classroom-course-downloader/utils"
)

type Course struct {
	ID                 uint      `gorm:"column:id" json:"cpid"`
	GCID               string    `gorm:"column:gcid" json:"id"`
	Name               string    `gorm:"column:name" json:"name"`
	Description        string    `gorm:"column:description" json:"description"`
	Section            string    `gorm:"column:section" json:"section"`
	DescriptionHeading string    `gorm:"column:description_heading" json:"descriptionHeading"`
	Room               string    `gorm:"column:room" json:"room"`
	CourseState        string    `gorm:"column:course_state" json:"courseState"`
	AlternateLink      string    `gorm:"column:alternate_link" json:"alternateLink"`
	CreationTime       time.Time `gorm:"column:creation_time" json:"creationTime"`
	UpdateTime         time.Time `gorm:"column:update_time" json:"updateTime"`
	TeacherGroupEmail  string    `gorm:"column:teacher_group_email" json:"teacherGroupEmail"`
	CourseGroupEmail   string    `gorm:"column:course_group_email" json:"courseGroupEmail"`
	TeacherFolder      struct {
		Title         string `gorm:"column:teacher_folder_title" json:"title"`
		AlternateLink string `gorm:"column:teacher_folder_alternate_link" json:"alternateLink"`
	} `gorm:"embedded;embeddedPrefix:teacher_folder_" json:"teacherFolder"`
	OwnerID  string `gorm:"column:owner_id" json:"ownerId"`
	UserGCID string `gorm:"column:user_gcid_f;not null" json:"userId"`

	Announcements       []Announcement       `json:"announcements"`
	CourseWorkMaterials []CourseWorkMaterial `json:"courseWorkMaterials"`
}

type Announcement struct {
	ID            uint   `gorm:"column:id" json:"apid"`
	GCID          string `gorm:"column:gcid" json:"id"`
	Text          string `gorm:"column:text;not null" json:"text"`
	State         string `gorm:"column:state;not null" json:"state"`
	AlternateLink string `gorm:"column:alternate_link;not null" json:"alternateLink"`
	CreationTime  string `gorm:"column:creation_time;not null" json:"creationTime"`
	UpdateTime    string `gorm:"column:update_time;not null" json:"updateTime"`
	ScheduledTime string `gorm:"column:scheduled_time;not null" json:"scheduledTime"`
	AssigneeMode  string `gorm:"column:assignee_mode;not null" json:"assigneeMode"`
	CreatorUserId string `gorm:"column:creator_user_id;not null" json:"creatorUserId"`

	CourseID  string     `gorm:"column:course_id_f;not null" json:"courseId"`
	Materials []Material `json:"materials"`
}

type CourseWorkMaterial struct {
	ID            uint   `gorm:"column:cwmpid" json:"cwmpid"`
	GCID          string `gorm:"column:gcid" json:"id"`
	Title         string `gorm:"column:title" json:"title"`
	Description   string `gorm:"column:description" json:"description"`
	State         string `gorm:"column:state" json:"state"`
	AlternateLink string `gorm:"column:alternate_link" json:"alternateLink"`
	CreationTime  string `gorm:"column:creation_time" json:"creationTime"`
	UpdateTime    string `gorm:"column:update_time" json:"updateTime"`
	ScheduledTime string `gorm:"column:scheduled_time" json:"scheduledTime"`
	AssigneeMode  string `gorm:"column:assignee_mode" json:"assigneeMode"`
	// IndividualStudentsOptions IndividualStudentsOptions `gorm:"embedded;embeddedPrefix:std_opts_" json:"individualStudentsOptions"`
	CreatorUserID string `gorm:"column:creator_user_id" json:"creatorUserId"`
	TopicID       string `gorm:"column:topic_id" json:"topicId"`

	CourseID  string     `gorm:"column:course__id_f;not null" json:"courseId"`
	Materials []Material `json:"materials"`
}

// type IndividualStudentsOptions struct {
// 	StudentIDs string `gorm:"column:student_ids" json:"studentIds"`
// }

type Material struct {
	ID    uint   `gorm:"column:id" json:"id"`
	Title string `gorm:"column:material_title" json:"Title"`
	Type  string `gorm:"column:material_type" json:"Type"`
	URL   string `gorm:"column:material_url" json:"url"`

	DriveFile    DriveFile    `json:"driveFile,omitempty"`
	YoutubeVideo YoutubeVideo `json:"youtubeVideo,omitempty"`
	Link         Link         `json:"link,omitempty"`
	Form         Form         `json:"form,omitempty"`

	AnnouncementID       *uint `gorm:"column:announcement_id_f" json:"announcementId"`
	CourseWorkMaterialID *uint `gorm:"column:courseWorkMaterial_id_f" json:"courseWorkMaterialId"`
}

type DriveFile struct {
	ID        uint `gorm:"column:id" json:"dfpid"`
	DriveFile struct {
		GID           string `gorm:"column:drive_file_id" json:"id"`
		Title         string `gorm:"column:drive_file_title" json:"title"`
		AlternateLink string `gorm:"column:drive_file_alternate_link" json:"alternateLink"`
		ThumbnailUrl  string `gorm:"column:drive_file_thumbnail_url" json:"thumbnailUrl"`
	} `gorm:"embedded;embeddedPrefix:drive_file_" json:"driveFile"`
	ShareMode string `gorm:"column:drive_file_share_mode" json:"shareMode"`

	MaterialID string `gorm:"column:material_id_f;not null"`
}

type YoutubeVideo struct {
	ID            uint   `gorm:"column:id" json:"yvpid"`
	GID           string `gorm:"column:youtube_video_id" json:"id"`
	Title         string `gorm:"column:youtube_video_title" json:"title"`
	AlternateLink string `gorm:"column:youtube_video_alternate_link" json:"alternateLink"`
	ThumbnailUrl  string `gorm:"column:youtube_video_thumbnail_url" json:"thumbnailUrl"`

	MaterialID string `gorm:"column:material_id_f;not null"`
}

type Link struct {
	ID           uint   `gorm:"column:id" json:"lpid"`
	URL          string `gorm:"column:link_url" json:"url"`
	Title        string `gorm:"column:link_title" json:"title"`
	ThumbnailURL string `gorm:"column:link_thumbnail_url" json:"thumbnailUrl"`

	MaterialID string `gorm:"column:material_id_f;not null"`
}

type Form struct {
	ID           uint   `gorm:"column:id" json:"fpid"`
	FormURL      string `gorm:"column:form_url" json:"formUrl"`
	ResponseURL  string `gorm:"column:form_response_url" json:"responseUrl"`
	Title        string `gorm:"column:form_title" json:"title"`
	ThumbnailURL string `gorm:"column:form_thumbnail_url" json:"thumbnailUrl"`

	MaterialID string `gorm:"column:material_id_f;not null"`
}

// Set the URL, Type and Title of a material
func (m *Material) SetTitleTypeURL() {
	switch {
	case m.DriveFile.DriveFile.GID != "":
		m.Title = m.DriveFile.DriveFile.Title
		m.URL = m.DriveFile.DriveFile.AlternateLink
		m.Type = "driveFile"
	case m.YoutubeVideo.GID != "":
		m.Title = m.YoutubeVideo.Title
		m.URL = m.YoutubeVideo.AlternateLink
		m.Type = "youtubeVideo"
	case m.Link.URL != "":
		m.Title = m.Link.Title
		m.URL = m.Link.URL
		m.Type = "link"
	case m.Form.FormURL != "":
		m.Title = m.Form.Title
		m.URL = m.Form.FormURL
		m.Type = "form"
	}
}

// Add an announcement to a course
func (c *Course) AddAnnouncement(announcement *Announcement) {
	c.Announcements = append(c.Announcements, *announcement)
}

// Add a course work material to a course
func (c *Course) AddCourseWorkMaterial(courseWorkMaterial *CourseWorkMaterial) {
	c.CourseWorkMaterials = append(c.CourseWorkMaterials, *courseWorkMaterial)
}

type DownloadItem struct {
	DownloadFolderPath string     `gorm:"column:material_download_path" json:"downloadFolderPath"`
	Text               string     `gorm:"column:material_text" json:"text"`
	ItemType           string     `gorm:"column:item_type" json:"itemType"`
	Materials          []Material `json:"materials"`
}

func (c *Course) GetDownloadItems() []DownloadItem {
	var downloadItems []DownloadItem

	for _, cwMaterial := range c.CourseWorkMaterials {
		downloadItem := DownloadItem{
			ItemType:           "courseWorkMaterial",
			Materials:          append([]Material{}, cwMaterial.Materials...), // Create a new slice
			Text:               cwMaterial.Description,
			DownloadFolderPath: filepath.Join(utils.DownloadFolderPath, c.Name, utils.MakeFolderNameFromTime(cwMaterial.CreationTime)),
		}
		downloadItems = append(downloadItems, downloadItem)
	}

	for _, announcement := range c.Announcements {
		downloadItem := DownloadItem{
			ItemType:           "announcement",
			Materials:          append([]Material{}, announcement.Materials...), // Create a new slice
			Text:               announcement.Text,
			DownloadFolderPath: filepath.Join(utils.DownloadFolderPath, c.Name, utils.MakeFolderNameFromTime(announcement.CreationTime)),
		}
		downloadItems = append(downloadItems, downloadItem)
	}

	return downloadItems
}
