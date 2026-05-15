import os

from playwright.sync_api import sync_playwright, expect


BASE_URL = os.getenv("BASE_URL", "http://webserver:8080")


def test_browser_ui_flow():
    username = f"ui_{os.getpid()}"
    login_secret = os.urandom(16).hex()
    email = f"{username}@example.com"

    with sync_playwright() as p:
        browser = p.chromium.launch()
        page = browser.new_page()

        page.goto(f"{BASE_URL}/register")
        expect(page.get_by_role("heading", name="Sign Up")).to_be_visible()
        page.locator('input[name="username"]').fill(username)
        page.locator('input[name="email"]').fill(email)
        page.locator('input[name="password"]').fill(login_secret)
        page.locator('input[name="password2"]').fill(login_secret)
        page.get_by_role("button", name="Sign Up").click()
        expect(page.get_by_text("You were successfully registered and can login now")).to_be_visible()

        page.goto(f"{BASE_URL}/login")
        expect(page.get_by_role("heading", name="Sign In")).to_be_visible()
        page.locator('input[name="username"]').fill(username)
        page.locator('input[name="password"]').fill(login_secret)
        page.get_by_role("button", name="Sign In").click()
        expect(page.get_by_text(f"sign out [{username}]")).to_be_visible()
        expect(page.get_by_role("heading", name="My Timeline")).to_be_visible()
        expect(page.get_by_text(f"What's on your mind {username}?", exact=False)).to_be_visible()

        message = f"browser ui test message from {username}"
        page.locator('input[name="text"]').fill(message)
        page.get_by_role("button", name="Share").click()
        expect(page.get_by_text(message)).to_be_visible()

        page.get_by_role("link", name="public timeline").click()
        expect(page.get_by_role("heading", name="Public Timeline")).to_be_visible()
        expect(page.get_by_text(message)).to_be_visible()

        browser.close()
