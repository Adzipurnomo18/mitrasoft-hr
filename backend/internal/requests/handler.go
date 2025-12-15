package requests

import (
	"bytes"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/xuri/excelize/v2"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateRequest(c *fiber.Ctx) error {
	val := c.Locals("userID")
	userID, ok := val.(int64)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid user session")
	}

	var req struct {
		Type      string `json:"type"`
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
		Reason    string `json:"reason"`
		// StartTime/EndTime could be separate if needed, assuming ISO8601 strings
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	start, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid start date format (RFC3339 required)"})
	}
	end, err := time.Parse(time.RFC3339, req.EndDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid end date format (RFC3339 required)"})
	}

	created, err := h.service.CreateRequest(c.Context(), userID, req.Type, start, end, req.Reason)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(created)
}

func (h *Handler) GetMyRequests(c *fiber.Ctx) error {
	val := c.Locals("userID")
	userID, ok := val.(int64)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid user session")
	}

	requests, err := h.service.GetMyRequests(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(requests)
}

func (h *Handler) GetPendingRequests(c *fiber.Ctx) error {
	// TODO: Verify permission (APPROVE_LEAVE or APPROVE_OVERTIME)
	// For now, assume protected route checks general access, but specific permission check is better.

	requests, err := h.service.GetPendingRequests(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(requests)
}

func (h *Handler) ApproveRequest(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	val := c.Locals("userID")
	approverID, ok := val.(int64)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid user session")
	}

	if err := h.service.ApproveRequest(c.Context(), id, approverID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Request approved"})
}

func (h *Handler) RejectRequest(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	val := c.Locals("userID")
	approverID, ok := val.(int64)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid user session")
	}

	var body struct {
		Reason string `json:"reason"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	if err := h.service.RejectRequest(c.Context(), id, approverID, body.Reason); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Request rejected"})
}

// GET /api/requests/summary?month=YYYY-MM
func (h *Handler) GetSummary(c *fiber.Ctx) error {
	month := c.Query("month")
	if month == "" {
		// Default to current month
		month = time.Now().Format("2006-01")
	}
	t, err := time.Parse("2006-01", month)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid month format, use YYYY-MM"})
	}
	from := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)

	s, err := h.service.GetSummaryBetween(c.Context(), from, to)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(s)
}

// GET /api/requests/summary/my?month=YYYY-MM
func (h *Handler) GetMySummary(c *fiber.Ctx) error {
	month := c.Query("month")
	if month == "" {
		month = time.Now().Format("2006-01")
	}
	t, err := time.Parse("2006-01", month)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid month format, use YYYY-MM"})
	}
	from := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)

	val := c.Locals("userID")
	userID, ok := val.(int64)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid user session")
	}

	s, err := h.service.GetMySummaryBetween(c.Context(), userID, from, to)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(s)
}

// GET /api/requests/processed?month=YYYY-MM
func (h *Handler) GetProcessedByMonth(c *fiber.Ctx) error {
	month := c.Query("month")
	if month == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "month query required, format YYYY-MM"})
	}
	t, err := time.Parse("2006-01", month)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid month format, use YYYY-MM"})
	}
	from := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)

	items, err := h.service.GetProcessedBetween(c.Context(), from, to)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(items)
}

// GET /api/requests/processed/export?month=YYYY-MM
func (h *Handler) ExportProcessedByMonth(c *fiber.Ctx) error {
	month := c.Query("month")
	if month == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "month query required, format YYYY-MM"})
	}
	t, err := time.Parse("2006-01", month)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid month format, use YYYY-MM"})
	}
	from := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)

	items, err := h.service.GetProcessedBetween(c.Context(), from, to)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Build XLSX with nice borders and ready to print
	f := excelize.NewFile()
	sheet := "Report"
	f.NewSheet(sheet)
	f.DeleteSheet("Sheet1")

	// Header styles
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#E5E7EB"}, Pattern: 1},
		Border:    []excelize.Border{{Type: "left", Color: "000000", Style: 1}, {Type: "right", Color: "000000", Style: 1}, {Type: "top", Color: "000000", Style: 1}, {Type: "bottom", Color: "000000", Style: 1}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	cellStyle, _ := f.NewStyle(&excelize.Style{
		Border:    []excelize.Border{{Type: "left", Color: "000000", Style: 1}, {Type: "right", Color: "000000", Style: 1}, {Type: "top", Color: "000000", Style: 1}, {Type: "bottom", Color: "000000", Style: 1}},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	dateStyle, _ := f.NewStyle(&excelize.Style{
		Border:    []excelize.Border{{Type: "left", Color: "000000", Style: 1}, {Type: "right", Color: "000000", Style: 1}, {Type: "top", Color: "000000", Style: 1}, {Type: "bottom", Color: "000000", Style: 1}},
		NumFmt:    22, // m/d/yy h:mm (Excel built-in)
		Alignment: &excelize.Alignment{Vertical: "center"},
	})

	// Set headers
	headers := []string{"ID", "Employee", "Type", "Start Date", "End Date", "Status", "Approver", "Updated At"}
	for i, hname := range headers {
		cell := fmt.Sprintf("%s1", string('A'+i))
		f.SetCellValue(sheet, cell, hname)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
	}
	// Column widths
	f.SetColWidth(sheet, "A", "A", 8)
	f.SetColWidth(sheet, "B", "B", 24)
	f.SetColWidth(sheet, "C", "C", 14)
	f.SetColWidth(sheet, "D", "E", 18)
	f.SetColWidth(sheet, "F", "F", 12)
	f.SetColWidth(sheet, "G", "G", 20)
	f.SetColWidth(sheet, "H", "H", 22)

	// Rows
	for idx, it := range items {
		row := idx + 2
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), it.ID)
		f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), cellStyle)

		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), it.UserName)
		f.SetCellStyle(sheet, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), cellStyle)

		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), it.Type)
		f.SetCellStyle(sheet, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), cellStyle)

		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), it.StartDate)
		f.SetCellStyle(sheet, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), dateStyle)

		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), it.EndDate)
		f.SetCellStyle(sheet, fmt.Sprintf("E%d", row), fmt.Sprintf("E%d", row), dateStyle)

		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), it.Status)
		f.SetCellStyle(sheet, fmt.Sprintf("F%d", row), fmt.Sprintf("F%d", row), cellStyle)

		approver := it.ApproverName
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), approver)
		f.SetCellStyle(sheet, fmt.Sprintf("G%d", row), fmt.Sprintf("G%d", row), cellStyle)

		f.SetCellValue(sheet, fmt.Sprintf("H%d", row), it.UpdatedAt)
		f.SetCellStyle(sheet, fmt.Sprintf("H%d", row), fmt.Sprintf("H%d", row), dateStyle)
	}

	// Freeze header row
	f.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		Split:       true,
		XSplit:      0,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	})

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to generate excel file")
	}

	filename := fmt.Sprintf("requests_%s.xlsx", month)
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	return c.SendStream(bytes.NewReader(buf.Bytes()))
}
