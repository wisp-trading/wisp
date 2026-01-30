package ui_test

import (
	"github.com/charmbracelet/lipgloss"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/wisp-trading/wisp/internal/ui"
	_ "github.com/wisp-trading/wisp/internal/ui/themes" // Register themes
)

var _ = Describe("Theme System", func() {
	Describe("Theme Registration", func() {
		It("should have at least 3 themes registered", func() {
			themes := ui.GetAvailableThemes()
			Expect(len(themes)).To(BeNumerically(">=", 3))
		})

		It("should include default, golden, and orange themes", func() {
			themes := ui.GetAvailableThemes()
			Expect(themes).To(ContainElement("default"))
			Expect(themes).To(ContainElement("golden"))
			Expect(themes).To(ContainElement("orange"))
		})
	})

	Describe("Theme Switching", func() {
		var originalColor lipgloss.Color

		BeforeEach(func() {
			// Store original color
			originalColor = ui.ColorPrimary
		})

		AfterEach(func() {
			// Always switch back to default after each test
			_ = ui.SetTheme("default")
		})

		Context("when switching to golden theme", func() {
			It("should update the current theme", func() {
				err := ui.SetTheme("golden")
				Expect(err).NotTo(HaveOccurred())

				theme := ui.GetCurrentTheme()
				Expect(theme.Name).To(Equal("golden"))
			})

			It("should change the color palette", func() {
				err := ui.SetTheme("golden")
				Expect(err).NotTo(HaveOccurred())

				Expect(ui.ColorPrimary).NotTo(Equal(originalColor))
			})

			It("should set ColorPrimary to gold (#FFD700)", func() {
				err := ui.SetTheme("golden")
				Expect(err).NotTo(HaveOccurred())

				expectedColor := lipgloss.Color("#FFD700")
				Expect(ui.ColorPrimary).To(Equal(expectedColor))
			})
		})

		Context("when switching to orange theme", func() {
			It("should set ColorPrimary to orange (#FF9900)", func() {
				err := ui.SetTheme("orange")
				Expect(err).NotTo(HaveOccurred())

				expectedColor := lipgloss.Color("#FF9900")
				Expect(ui.ColorPrimary).To(Equal(expectedColor))
			})
		})

		Context("when switching back to default theme", func() {
			It("should restore the original color", func() {
				// Switch to another theme first
				err := ui.SetTheme("golden")
				Expect(err).NotTo(HaveOccurred())

				// Switch back to default
				err = ui.SetTheme("default")
				Expect(err).NotTo(HaveOccurred())

				Expect(ui.ColorPrimary).To(Equal(originalColor))
			})
		})
	})

	Describe("Invalid Theme Handling", func() {
		It("should return an error when setting a nonexistent theme", func() {
			err := ui.SetTheme("nonexistent")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not found"))
		})
	})

	Describe("GetTheme", func() {
		Context("when getting golden theme", func() {
			It("should return the theme object", func() {
				theme := ui.GetTheme("golden")
				Expect(theme).NotTo(BeNil())
			})

			It("should have the correct name", func() {
				theme := ui.GetTheme("golden")
				Expect(theme.Name).To(Equal("golden"))
			})

			It("should have the correct primary color", func() {
				theme := ui.GetTheme("golden")
				Expect(theme.Colors.Primary).To(Equal("#FFD700"))
			})
		})

		Context("when getting a nonexistent theme", func() {
			It("should return nil", func() {
				theme := ui.GetTheme("nonexistent")
				Expect(theme).To(BeNil())
			})
		})
	})

	Describe("Style Rebuilding", func() {
		AfterEach(func() {
			// Clean up by switching back to default
			_ = ui.SetTheme("default")
		})

		It("should rebuild styles with new colors when theme changes", func() {
			// Switch to orange theme
			err := ui.SetTheme("orange")
			Expect(err).NotTo(HaveOccurred())

			// Verify ColorPrimary is now orange
			expectedColor := lipgloss.Color("#FF9900")
			Expect(ui.ColorPrimary).To(Equal(expectedColor))
		})
	})
})
