package globals

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/jlaffaye/ftp"
)

func (gs *GlobalState) FetchReportFiles() (err error) {
	ipa := gs.Cfg.Plc.IPAddress + ":21"

	log.Println("📄 Dialing Ftp Server at", ipa)

	c, err := ftp.Dial(ipa, ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		return fmt.Errorf("📄 Error dialing FTP at %s: %w", ipa, err)
	}

	err = c.Login("Admin", "admin")
	if err != nil {
		return fmt.Errorf("📄 Error logging in to FTP: %w", err)
	}

	//check if target for report data exists, otherwise create it
	if gs.Report.All == nil {
		gs.Report.All = make([][]ReportCsvLine, gs.Cfg.Plc.SWSettings.Raumanzahl)
	}

	sharedFolderAvailable := sharedFolderAvailable()

	for i := 1; i <= gs.Cfg.Plc.SWSettings.Raumanzahl; i++ {
		fileName := fmt.Sprintf("/sdcard/Reporte/Rohdaten/logData_Raum%d.csv", i)
		r, err := c.Retr(fileName)
		if err != nil {
			return fmt.Errorf("📄 Error retrieving file: %s, %w", fileName, err)
		}

		buf, err := io.ReadAll(r)
		if err != nil {
			return fmt.Errorf("📄 Error reading file: %s, %w", fileName, err)
		}

		if sharedFolderAvailable {
			writeToSharedFolder(buf, i)
		}

		var lines []ReportCsvLine

		gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
			r := csv.NewReader(in)
			r.Comma = ';'
			r.FieldsPerRecord = -1
			return r // Allows use semicolon as delimiter
		})

		if err := gocsv.UnmarshalBytes(buf, &lines); err != nil {
			return fmt.Errorf("📄 Error unmarshalling CSV: %w", err)
		}

		// Trim null characters from Datum field for each line
		for j := range lines {
			lines[j].Datum = strings.ReplaceAll(lines[j].Datum, "\u0000", "")
		}

		// Save to global state, Index starts with 0
		gs.Report.All[i-1] = lines

		if err := r.Close(); err != nil {

			return fmt.Errorf("📄 Error closing reader: %w", err)
		}

	}

	// Debug
	log.Println("📄 Report.All length:", len(gs.Report.All))
	for i, r := range gs.Report.All {
		log.Printf("📄 Report.All[%d] length: %d\n", i, len(r))
		if len(r) > 0 {
			log.Printf("📄 First line: %+v\n", r[0])
		}
	}

	if err := c.Quit(); err != nil {
		return fmt.Errorf("📄 Error quitting FTP: %w", err)
	}
	return nil
}

func sharedFolderAvailable() bool {
	sfolder := "/shared"
	// Check if folder exists

	fstat, err := os.Stat(sfolder)
	if err != nil {
		log.Println("📄 Shared folder not available:", err)
		return false
	}
	log.Println("📄 Shared folder available:", sfolder, "Info:", fstat.Mode().Perm())

	return true
}

func writeToSharedFolder(buf []byte, fileIdx int) {

	sharedFilePath := fmt.Sprintf("/shared/logData_Raum%d.csv", fileIdx)
	err := os.WriteFile(sharedFilePath, buf, 0644)
	if err != nil {
		log.Println("📄 Error writing to shared folder:", err)
	} else {
		log.Println("📄 Successfully wrote to shared folder:", sharedFilePath)
	}
}
