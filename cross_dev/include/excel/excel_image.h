#ifndef EXCEL_IMAGE_H
#define EXCEL_IMAGE_H

#include <string>

namespace excel {

/**
 * Inject a letterhead image at the top of the first worksheet in an existing .xlsx file.
 * Call this AFTER the Excel file has been saved.
 *
 * @param xlsxPath Path to the .xlsx file (must exist and be a valid xlsx)
 * @param imagePath Path to the PNG image (e.g. letter_head.png)
 * @param widthInches Display width in inches (default 7.0, ~full letter width)
 * @param heightInches Display height in inches (default 1.5, typical letterhead strip)
 * @return true on success, false on failure (file not found, invalid xlsx, etc.)
 */
bool injectLetterhead(const std::string& xlsxPath, const std::string& imagePath,
                      double widthInches = 7.0, double heightInches = 1.5);

/**
 * Resolve path to letter_head.png (same candidates as template).
 */
std::string resolveLetterheadPath(const std::string& filename = "letter_head.png");

} // namespace excel

#endif // EXCEL_IMAGE_H
