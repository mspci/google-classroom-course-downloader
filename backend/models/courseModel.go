package models

import (
	"time"

	"github.com/mspcix/google-classroom-downloader/utils"
)

type Course struct {
	ID                 string    `gorm:"primaryKey" json:"id"`
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

	Announcements []Announcement `json:"announcements"`
	OwnerID       string         `gorm:"column:owner_id" json:"ownerId"`
	UserGCID      string         `gorm:"column:user_gcid_f;not null" json:"userId"`
}

type Announcement struct {
	ID            string `gorm:"primaryKey" json:"id"`
	Text          string `gorm:"column:text;not null" json:"text"`
	State         string `gorm:"column:state;not null" json:"state"`
	AlternateLink string `gorm:"column:alternate_link;not null" json:"alternateLink"`
	CreationTime  string `gorm:"column:creation_time;not null" json:"creationTime"`
	UpdateTime    string `gorm:"column:update_time;not null" json:"updateTime"`
	ScheduledTime string `gorm:"column:scheduled_time;not null" json:"scheduledTime"`
	AssigneeMode  string `gorm:"column:assignee_mode;not null" json:"assigneeMode"`
	CreatorUserId string `gorm:"column:creator_user_id;not null" json:"creatorUserId"`

	Materials []Material `json:"materials"`
	CourseID  string     `gorm:"column:course_id;not null" json:"courseId"`
}

type Material struct {
	ID    uint   `gorm:"primaryKey" json:"id"`
	Title string `gorm:"column:material_title" json:"Title"`
	Type  string `gorm:"column:material_type" json:"Type"`
	URL   string `gorm:"column:material_url" json:"url"`

	DriveFile    DriveFile    `json:"driveFile,omitempty"`
	YoutubeVideo YoutubeVideo `json:"youtubeVideo,omitempty"`
	Link         Link         `json:"link,omitempty"`
	Form         Form         `json:"form,omitempty"`

	AnnouncementID string `gorm:"column:announcement_id;not null" json:"announcementId"`
}

// type Material struct {
// 	ID    uint   `gorm:"primaryKey" json:"id"`

// 	Title string `gorm:"column:material_title" json:"Title"`
// 	Type  string `gorm:"column:material_type" json:"Type"`

// 	MaterialDetail MaterialDetail `json:"materialDetail,omitempty"`

// 	AnnouncementID string `gorm:"column:announcement_id;not null" json:"announcementId"`
// }

// type MaterialDetail interface {
// 	isMaterialDetail()
// 	GetTitle() string
// 	GetType() string
// }

type DriveFile struct {
	DriveFile struct {
		ID            string `gorm:"column:drive_file_id;primaryKey" json:"id"`
		Title         string `gorm:"column:drive_file_title" json:"title"`
		AlternateLink string `gorm:"column:drive_file_alternate_link" json:"alternateLink"`
		ThumbnailUrl  string `gorm:"column:drive_file_thumbnail_url" json:"thumbnailUrl"`
	} `gorm:"embedded;embeddedPrefix:drive_file_" json:"driveFile"`
	ShareMode string `gorm:"column:drive_file_share_mode" json:"shareMode"`

	MaterialID string `gorm:"column:material_id;not null"`
}

// func (DriveFile) isMaterialDetail() {}
// func (d DriveFile) GetTitle() string {
// 	return d.DriveFile.Title
// }
// func (d DriveFile) GetType() string {
// 	return "driveFile"
// }

type YoutubeVideo struct {
	ID            string `gorm:"column:youtube_video_id;primaryKey" json:"id"`
	Title         string `gorm:"column:youtube_video_title" json:"title"`
	AlternateLink string `gorm:"column:youtube_video_alternate_link" json:"alternateLink"`
	ThumbnailUrl  string `gorm:"column:youtube_video_thumbnail_url" json:"thumbnailUrl"`

	MaterialID string `gorm:"column:material_id;not null"`
}

// func (YoutubeVideo) isMaterialDetail() {}
// func (y YoutubeVideo) GetTitle() string {
// 	return y.Title
// }
// func (y YoutubeVideo) GetType() string {
// 	return "youtubeVideo"
// }

type Link struct {
	ID           uint   `gorm:"column:link_id;primaryKey"`
	URL          string `gorm:"column:link_url" json:"url"`
	Title        string `gorm:"column:link_title" json:"title"`
	ThumbnailURL string `gorm:"column:link_thumbnail_url" json:"thumbnailUrl"`

	MaterialID string `gorm:"column:material_id;not null"`
}

// func (Link) isMaterialDetail() {}
// func (l Link) GetTitle() string {
// 	return l.Title
// }
// func (l Link) GetType() string {
// 	return "link"
// }

type Form struct {
	ID           uint   `gorm:"column:form_id;primaryKey"`
	FormURL      string `gorm:"column:form_url" json:"formUrl"`
	ResponseURL  string `gorm:"column:form_response_url" json:"responseUrl"`
	Title        string `gorm:"column:form_title" json:"title"`
	ThumbnailURL string `gorm:"column:form_thumbnail_url" json:"thumbnailUrl"`

	MaterialID string `gorm:"column:material_id;not null"`
}

// func (Form) isMaterialDetail() {}
// func (f Form) GetTitle() string {
// 	return f.Title
// }
// func (f Form) GetType() string {
// 	return "form"
// }

// func (m *Material) SetTitleAndTypeFromDetail() {
// 	if m.MaterialDetail != nil {
// 		m.Title = m.MaterialDetail.GetTitle()
// 		m.Type = m.MaterialDetail.GetType()
// 	}
// }

// Set the URL, Type and Title of a material
func (m *Material) SetTitleTypeURL() {
	switch {
	case m.DriveFile.DriveFile.ID != "":
		m.Title = utils.RemoveInvalidChars(m.DriveFile.DriveFile.Title)
		m.URL = m.DriveFile.DriveFile.AlternateLink
		m.Type = "driveFile"
	case m.YoutubeVideo.ID != "":
		m.Title = utils.RemoveInvalidChars(m.YoutubeVideo.Title)
		m.URL = m.YoutubeVideo.AlternateLink
		m.Type = "youtubeVideo"
	case m.Link.URL != "":
		m.Title = utils.RemoveInvalidChars(m.Link.Title)
		m.URL = m.Link.URL
		m.Type = "link"
	case m.Form.FormURL != "":
		m.Title = utils.RemoveInvalidChars(m.Form.Title)
		m.URL = m.Form.FormURL
		m.Type = "form"
	}
}

// Set the material of an announcement
func (a *Announcement) SetMaterial(material *Material) {
	a.Materials = append(a.Materials, *material)
}

// Add an announcement to a course
func (c *Course) AddAnnouncement(announcement *Announcement) {
	c.Announcements = append(c.Announcements, *announcement)
}
