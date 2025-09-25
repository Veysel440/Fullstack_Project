import { test, expect } from "@playwright/test";

test("login as admin and CRUD", async ({ page }) => {
    await page.goto("/");
    await page.fill('input[name="email"]', "admin@example.com");
    await page.fill('input[name="password"]', "admin123");
    await page.click('button:has-text("Login")');
    await expect(page.getByText("role: admin")).toBeVisible();

    await page.fill('input[placeholder="Name"]', "e2e item");
    await page.fill('input[placeholder="Price"]', "12.34");
    await page.click('button:has-text("Add")');
    await expect(page.getByText("e2e item")).toBeVisible();

    await page.getByRole("button", { name: "Delete" }).first().click();
});

test("login as user hides Delete", async ({ page }) => {
    await page.goto("/");
    await page.fill('input[name="email"]', "user@example.com");
    await page.fill('input[name="password"]', "user123");
    await page.click('button:has-text("Login")');
    await expect(page.getByText("role: user")).toBeVisible();
    await expect(page.getByRole("button", { name: "Delete" })).toHaveCount(0);
});
