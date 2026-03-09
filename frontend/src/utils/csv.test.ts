import { describe, it, expect } from "vitest";
import { parseCsv } from "./csv";

describe("parseCsv", () => {
  it("parses basic CSV with headers", () => {
    const input = "date,amount,category\n2026-01-15,42.50,food\n2026-01-16,15.00,transport";
    const result = parseCsv(input);
    expect(result.headers).toEqual(["date", "amount", "category"]);
    expect(result.rows).toHaveLength(2);
    expect(result.rows[0]).toEqual(["2026-01-15", "42.50", "food"]);
    expect(result.rows[1]).toEqual(["2026-01-16", "15.00", "transport"]);
  });

  it("handles quoted fields with commas", () => {
    const input = 'date,description,amount\n2026-01-15,"Coffee, Donuts",5.00';
    const result = parseCsv(input);
    expect(result.rows[0]).toEqual(["2026-01-15", "Coffee, Donuts", "5.00"]);
  });

  it("handles quoted fields with newlines", () => {
    const input = 'date,description,amount\n2026-01-15,"Multi\nline",5.00';
    const result = parseCsv(input);
    expect(result.rows[0]).toEqual(["2026-01-15", "Multi\nline", "5.00"]);
  });

  it("handles escaped quotes inside quoted fields", () => {
    const input = 'name,value\n"""Hello""",42';
    const result = parseCsv(input);
    expect(result.rows[0]).toEqual(['"Hello"', "42"]);
  });

  it("trims whitespace from unquoted fields", () => {
    const input = "a , b , c \n 1 , 2 , 3 ";
    const result = parseCsv(input);
    expect(result.headers).toEqual(["a", "b", "c"]);
    expect(result.rows[0]).toEqual(["1", "2", "3"]);
  });

  it("returns empty rows for empty input", () => {
    const result = parseCsv("");
    expect(result.headers).toEqual([]);
    expect(result.rows).toEqual([]);
  });

  it("returns headers only when no data rows exist", () => {
    const result = parseCsv("date,amount,category");
    expect(result.headers).toEqual(["date", "amount", "category"]);
    expect(result.rows).toEqual([]);
  });

  it("skips empty lines", () => {
    const input = "a,b\n1,2\n\n3,4\n";
    const result = parseCsv(input);
    expect(result.rows).toHaveLength(2);
    expect(result.rows[0]).toEqual(["1", "2"]);
    expect(result.rows[1]).toEqual(["3", "4"]);
  });

  it("handles Windows-style line endings", () => {
    const input = "a,b\r\n1,2\r\n3,4";
    const result = parseCsv(input);
    expect(result.headers).toEqual(["a", "b"]);
    expect(result.rows).toHaveLength(2);
  });
});
