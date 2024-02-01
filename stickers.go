package telebot

import (
	"encoding/json"
	"strconv"
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

type InputSticker struct {
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
// *File in InputStickers must be uploaded first!
func (b *Bot) CreateStickerSet(to Recipient, format string, inputs []InputSticker, s StickerSet) error {
	inputStickers, _ := json.Marshal(inputs)
	params := map[string]string{
		"stickers":       string(inputStickers),
		"user_id":        to.Recipient(),
		"sticker_type":   s.Type,
		"sticker_format": format,
		"name":           s.Name,
		"title":          s.Title,
		// "emojis":       s.Emojis,
		// "contains_masks":   strconv.FormatBool(s.ContainsMasks),
		"needs_repainting": strconv.FormatBool(s.NeedsRepainting),
	}

	// if s.MaskPosition != nil {
	// 	data, _ := json.Marshal(&s.MaskPosition)
	// 	params["mask_position"] = string(data)
	// }

	_, err := b.Raw("createNewStickerSet", params)
	return err
}

// AddSticker adds a new sticker to the existing sticker set.
// For fields in StickerSet, only Name is required.
func (b *Bot) AddSticker(to Recipient, input InputSticker, s StickerSet) error {
	inputSticker, _ := json.Marshal(input)
	params := map[string]string{
		"sticker": string(inputSticker),
		"user_id": to.Recipient(),
		"name":    s.Name,
	}

	_, err := b.Raw("addStickerToSet", params)
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
func (b *Bot) SetStickerSetTitle(to Recipient, title string, s StickerSet) error {
	_, err := b.Raw("setStickerSetTitle", map[string]string{"name": s.Name, "title": title})
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
