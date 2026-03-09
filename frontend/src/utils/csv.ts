export interface CsvResult {
  headers: string[];
  rows: string[][];
}

export function parseCsv(input: string): CsvResult {
  if (input.trim() === "") {
    return { headers: [], rows: [] };
  }

  const lines = parseLines(input);
  if (lines.length === 0) {
    return { headers: [], rows: [] };
  }

  const firstLine = lines[0];
  if (!firstLine) {
    return { headers: [], rows: [] };
  }
  const headers = firstLine.map((h) => h.trim());
  const rows = lines
    .slice(1)
    .filter((row) => row.some((cell) => cell.trim() !== ""))
    .map((row) => row.map((cell) => cell.trim()));

  return { headers, rows };
}

function parseLines(input: string): string[][] {
  const results: string[][] = [];
  let current: string[] = [];
  let field = "";
  let inQuotes = false;
  let idx = 0;

  while (idx < input.length) {
    const char = input[idx];

    if (inQuotes) {
      if (char === '"') {
        if (idx + 1 < input.length && input[idx + 1] === '"') {
          field += '"';
          idx += 2;
          continue;
        }
        inQuotes = false;
        idx++;
        continue;
      }
      field += char;
      idx++;
      continue;
    }

    if (char === '"') {
      inQuotes = true;
      idx++;
      continue;
    }

    if (char === ",") {
      current.push(field);
      field = "";
      idx++;
      continue;
    }

    if (char === "\r" && idx + 1 < input.length && input[idx + 1] === "\n") {
      current.push(field);
      field = "";
      results.push(current);
      current = [];
      idx += 2;
      continue;
    }

    if (char === "\n") {
      current.push(field);
      field = "";
      results.push(current);
      current = [];
      idx++;
      continue;
    }

    field += char;
    idx++;
  }

  if (field !== "" || current.length > 0) {
    current.push(field);
    results.push(current);
  }

  return results;
}
