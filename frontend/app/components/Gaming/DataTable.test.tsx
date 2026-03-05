/**
 * @vitest-environment jsdom
 */
import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, test } from "vitest";
import DataTable, { DataTableCell, formatUnixTimestamp } from "./DataTable";

afterEach(() => {
  cleanup();
});

describe("DataTable", () => {
  test("renders headers", () => {
    render(
      <DataTable headers={["A", "B", "C"]}>
        <tr>
          <DataTableCell>1</DataTableCell>
          <DataTableCell>2</DataTableCell>
          <DataTableCell>3</DataTableCell>
        </tr>
      </DataTable>,
    );
    expect(screen.getByText("A")).toBeDefined();
    expect(screen.getByText("B")).toBeDefined();
    expect(screen.getByText("C")).toBeDefined();
  });

  test("renders body cells", () => {
    render(
      <DataTable headers={["H"]}>
        <tr>
          <DataTableCell>cell content</DataTableCell>
        </tr>
      </DataTable>,
    );
    expect(screen.getByText("cell content")).toBeDefined();
  });

  test("renders multiple rows", () => {
    render(
      <DataTable headers={["Name"]}>
        <tr>
          <DataTableCell>Alice</DataTableCell>
        </tr>
        <tr>
          <DataTableCell>Bob</DataTableCell>
        </tr>
      </DataTable>,
    );
    expect(screen.getByText("Alice")).toBeDefined();
    expect(screen.getByText("Bob")).toBeDefined();
  });
});

describe("formatUnixTimestamp", () => {
  test("formats timestamp correctly", () => {
    // 2026-03-01 12:30 JST (UTC+9) = 2026-03-01 03:30 UTC
    const timestamp = Date.UTC(2026, 2, 1, 3, 30, 0) / 1000;
    const result = formatUnixTimestamp(timestamp);
    // Result depends on local timezone; just check the format pattern
    expect(result).toMatch(/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}$/);
  });

  test("pads single-digit months and days", () => {
    // Use a date where month and day are single digits
    const timestamp = Date.UTC(2026, 0, 5, 0, 0, 0) / 1000;
    const result = formatUnixTimestamp(timestamp);
    expect(result).toMatch(/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}$/);
  });
});
