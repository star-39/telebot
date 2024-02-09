package telebot

import (
	"encoding/json"
	"strconv"
	"strings"
)

type StickerSetType = string

const (
	StickerRegular     = "regular"
	StickerMask        = "mask"
	StickerCustomEmoji = "custom_emoji"
)

// StickerSet represents a sticker set.
type StickerSet struct {
	Type      StickerSetType `json:"sticker_type"`
	Name      string         `json:"name"`
	Title     string         `json:"title"`
	Animated  bool           `json:"is_animated"`
	Video     bool           `json:"is_video"`
	Stickers  []Sticker      `json:"stickers"`
	Thumbnail *Photo         `json:"thumb"`
	// PNG             *File          `json:"png_sticker"`
	// TGS             *File          `json:"tgs_sticker"`
	// WebM            *File          `json:"webm_sticker"`
	Emojis          string        `json:"emojis"`
	ContainsMasks   bool          `json:"contains_masks"` // FIXME: can be removed
	MaskPosition    *MaskPosition `json:"mask_position"`
	NeedsRepainting bool          `json:"needs_repainting"`
}

func (ss StickerSet) Format() string {
	if ss.Video {
		return "video"
	} else {
		return "static"
	}
}

type InputSticker struct {
	//if starts with file:// , treat it as local file, otherwise, fileID
	Sticker string   `json:"sticker"`
	Emojis  []string `json:"emoji_list"`
	// MaskPosition MaskPosition `json:"mask_position"`
	Keywords []string `json:"keywords"`
}

// MaskPosition describes the position on faces where
// a mask should be placed by default.
type MaskPosition struct {
	Feature MaskFeature `json:"point"`
	XShift  float32     `json:"x_shift"`
	YShift  float32     `json:"y_shift"`
	Scale   float32     `json:"scale"`
}

// MaskFeature defines sticker mask position.
type MaskFeature string

const (
	FeatureForehead MaskFeature = "forehead"
	FeatureEyes     MaskFeature = "eyes"
	FeatureMouth    MaskFeature = "mouth"
	FeatureChin     MaskFeature = "chin"
)

// UploadSticker uploads a sticker file with a sticker for later use.
func (b *Bot) UploadSticker(to Recipient, format string, sticker *File) (*File, error) {
	files := map[string]File{
		"sticker": *sticker,
	}
	params := map[string]string{
		"user_id":        to.Recipient(),
		"sticker_format": format,
	}

	data, err := b.sendFiles("uploadStickerFile", files, params)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Result File
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, wrapError(err)
	}
	return &resp.Result, nil
}

// StickerSet returns a sticker set on success.
func (b *Bot) StickerSet(name string) (*StickerSet, error) {
	data, err := b.Raw("getStickerSet", map[string]string{"name": name})
	if err != nil {
		return nil, err
	}

	var resp struct {
		Result *StickerSet
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, wrapError(err)
	}
	return resp.Result, nil
}

// CreateStickerSet creates a new sticker set.
// StickerSet should include Type, Title, Name. Format will be guessed automatically.
// If InputFile in InputSticker starts with file:// , treat as local file.
func (b *Bot) CreateStickerSet(to Recipient, inputs []InputSticker, ss StickerSet) error {
	var hasLocalFile bool
	stickerFilesMap := make(map[string]File)
	for index, input := range inputs {
		// Upload as attach://
		if strings.HasPrefix(input.Sticker, "file://") {
			filePath := strings.TrimPrefix(input.Sticker, "file://")
			fileIdentifier := "sticker" + strconv.Itoa(index)
			inputs[index].Sticker = "attach://" + fileIdentifier
			stickerFilesMap[fileIdentifier] = File{FileLocal: filePath}
			hasLocalFile = true
		}
	}
	inputStickers, _ := json.Marshal(inputs)
	params := map[string]string{
		"stickers":       string(inputStickers),
		"user_id":        to.Recipient(),
		"sticker_type":   ss.Type,
		"sticker_format": ss.Format(),
		"name":           ss.Name,
		"title":          ss.Title,
	}

	var err error
	if hasLocalFile {
		_, err = b.sendFiles("createNewStickerSet", stickerFilesMap, params)
	} else {
		_, err = b.Raw("createNewStickerSet", params)
	}
	return err
}

// AddSticker adds a new sticker to the existing sticker set.
// StickerSet only requires Name.
// If InputFile in InputSticker starts with file:// , treat as local file.
func (b *Bot) AddSticker(to Recipient, input InputSticker, ss StickerSet) error {
	var hasLocalFile bool
	stickerFilesMap := make(map[string]File)
	if strings.HasPrefix(input.Sticker, "file://") {
		filePath := strings.TrimPrefix(input.Sticker, "file://")
		fileIdentifier := "sticker00"
		input.Sticker = "attach://" + fileIdentifier
		stickerFilesMap[fileIdentifier] = File{FileLocal: filePath}
		hasLocalFile = true
	}
	inputSticker, _ := json.Marshal(input)
	params := map[string]string{
		"sticker": string(inputSticker),
		"user_id": to.Recipient(),
		"name":    ss.Name,
	}

	var err error
	if hasLocalFile {
		_, err = b.sendFiles("addStickerToSet", stickerFilesMap, params)
	} else {
		_, err = b.Raw("addStickerToSet", params)
	}
	return err
}

// SetStickerPosition moves a sticker in set to a specific position.
func (b *Bot) SetStickerPosition(sticker string, position int) error {
	params := map[string]string{
		"sticker":  sticker,
		"position": strconv.Itoa(position),
	}

	_, err := b.Raw("setStickerPositionInSet", params)
	return err
}

// DeleteSticker deletes a sticker from a set created by the bot.
func (b *Bot) DeleteSticker(sticker string) error {
	_, err := b.Raw("deleteStickerFromSet", map[string]string{"sticker": sticker})
	return err

}

// SetStickerSetThumb sets a thumbnail of the sticker set.
//
// .WEBP or .PNG image with the thumbnail, must be up to 128 kilobytes in size and have a width and height of exactly 100px
// .TGS animation with a thumbnail up to 32 kilobytes in size
// WEBM video with the thumbnail up to 32 kilobytes in size
func (b *Bot) SetStickerSetThumbnail(to Recipient, file File, s StickerSet) error {
	files := make(map[string]File)
	files["thumbnail"] = file

	params := map[string]string{
		"name":    s.Name,
		"user_id": to.Recipient(),
	}

	_, err := b.sendFiles("setStickerSetThumbnail", files, params)
	return err
}

// Use this method to set the title of a created sticker set. Returns True on success.
func (b *Bot) SetStickerSetTitle(to Recipient, title string, name string) error {
	_, err := b.Raw("setStickerSetTitle", map[string]string{"name": name, "title": title})
	return err
}

// Use this method to change the list of emoji assigned to a regular or custom emoji sticker.
// The sticker must belong to a sticker set created by the bot.
func (b *Bot) SetStickerEmojiList(to Recipient, sticker string, emojis []string) error {
	emojiList, _ := json.Marshal(emojis)
	params := map[string]string{
		"sticker":    sticker,
		"emoji_list": string(emojiList),
	}

	_, err := b.Raw("setStickerEmojiList", params)
	return err
}

// CustomEmojiStickers returns the information about custom emoji stickers by their ids.
func (b *Bot) CustomEmojiStickers(ids []string) ([]Sticker, error) {
	data, _ := json.Marshal(ids)

	params := map[string]string{
		"custom_emoji_ids": string(data),
	}

	data, err := b.Raw("getCustomEmojiStickers", params)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Result []Sticker
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, wrapError(err)
	}
	return resp.Result, nil
}
