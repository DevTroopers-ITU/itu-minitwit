import os

from playwright.sync_api import sync_playwright


BASE_URL = os.getenv("BASE_URL", "http://127.0.0.1:8080")


def test_browser_ui_flow():
    username = f"ui_{os.getpid()}"
    login_secret = os.urandom(16).hex()
    email = f"{username}@example.com"

    with sync_playwright() as p:
        browser = p.chromium.launch()
        page = browser.new_page()

        page.goto(f"{BASE_URL}/register")
        assert page.get_by_role("heading", name="Sign Up").is_visible()
        page.locator('input[name="username"]').fill(username)
        page.locator('input[name="email"]').fill(email)
        page.locator('input[name="password"]').fill(login_secret)
        page.locator('input[name="password2"]').fill(login_secret)
        page.get_by_role("button", name="Sign Up").click()
        assert page.get_by_text("You were successfully registered and can login now").is_visible()

        page.goto(f"{BASE_URL}/login")
        assert page.get_by_role("heading", name="Sign In").is_visible()
        page.locator('input[name="username"]').fill(username)
        page.locator('input[name="password"]').fill(login_secret)
        page.get_by_role("button", name="Sign In").click()
        assert page.get_by_text(f"sign out [{username}]").is_visible()
        assert page.get_by_role("heading", name="My Timeline").is_visible()
        assert page.get_by_text(f"What's on your mind {username}?").is_visible()

        message = f"browser ui test message from {username}"
        page.locator('input[name="text"]').fill(message)
        page.get_by_role("button", name="Share").click()
        assert page.get_by_text(message).is_visible()

        page.get_by_role("link", name="public timeline").click()
        assert page.get_by_role("heading", name="Public Timeline").is_visible()
        assert page.get_by_text(message).is_visible()

        browser.close()
