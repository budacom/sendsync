package cmd_test

import (
	. "sendsync/cmd"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Common", func() {
	Describe("UpdateTemplateFromJson", func() {
		Context("When Json has more keys that needed", func() {
			It("Should load only known keys", func() {
				template := Template{}
				template.UpdateTemplateFromJson(`{"Generation":"Foo", "NotAKey":"Bar"}`)
				Expect(template).To(Equal(Template{Generation: "Foo"}))
			})
		})

		Context("When Json does not have known keys", func() {
			It("Should load an empty template", func() {
				template := Template{}
				template.UpdateTemplateFromJson(`{"random":"Foo", "NotAKey":"Bar"}`)
				Expect(template).To(Equal(Template{}))
			})
		})
	})

	Describe("FindActiveVersion", func() {
		It("Finds version that has a 1 on Active template", func() {
			template := Template{
				Versions: []Version{
					{Active: 0},
					{Active: 0},
					{Active: 1, Name: "Shrek"},
					{Active: 0},
				},
			}
			Expect(template.FindActiveVersion()).To(Equal(&Version{Active: 1, Name: "Shrek"}))
		})

		Context("When there is no active template", func() {
			It("Returns nil", func() {
				template := Template{
					Versions: []Version{
						{Active: 0},
						{Active: 0},
						{Active: 0},
						{Active: 0},
					},
				}
				var nilVersion *Version
				Expect(template.FindActiveVersion()).To(Equal(nilVersion))
			})
		})
	})
})
