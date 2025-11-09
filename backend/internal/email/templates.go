package email

import "fmt"

func renderWelcomeEmail(firstName string) string {
	name := firstName
	if name == "" {
		name = "there"
	}

	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Welcome to SaaS Blueprint</title>
</head>
<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;">
    <table width="100%%" cellpadding="0" cellspacing="0" style="background-color: #f7fafc; padding: 40px 20px;">
        <tr>
            <td align="center">
                <table width="600" cellpadding="0" cellspacing="0" style="background-color: #ffffff; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
                    <!-- Header -->
                    <tr>
                        <td style="padding: 40px 40px 20px 40px; text-align: center;">
                            <h1 style="margin: 0; color: #2d3748; font-size: 28px; font-weight: 700;">
                                Welcome to SaaS Blueprint!
                            </h1>
                        </td>
                    </tr>

                    <!-- Content -->
                    <tr>
                        <td style="padding: 0 40px 40px 40px;">
                            <p style="color: #4a5568; font-size: 16px; line-height: 1.6; margin: 0 0 20px 0;">
                                Hi %s,
                            </p>
                            <p style="color: #4a5568; font-size: 16px; line-height: 1.6; margin: 0 0 20px 0;">
                                Thank you for joining SaaS Blueprint! We're excited to have you on board.
                            </p>
                            <p style="color: #4a5568; font-size: 16px; line-height: 1.6; margin: 0 0 20px 0;">
                                You now have access to all our features. Here are some things you can do:
                            </p>
                            <ul style="color: #4a5568; font-size: 16px; line-height: 1.8; margin: 0 0 20px 0; padding-left: 20px;">
                                <li>Complete your profile</li>
                                <li>Explore the dashboard</li>
                                <li>Invite team members</li>
                                <li>Get started with your first project</li>
                            </ul>
                            <p style="color: #4a5568; font-size: 16px; line-height: 1.6; margin: 0 0 30px 0;">
                                If you have any questions, feel free to reach out to our support team.
                            </p>

                            <!-- CTA Button -->
                            <table width="100%%" cellpadding="0" cellspacing="0">
                                <tr>
                                    <td align="center">
                                        <a href="http://localhost:3000/dashboard"
                                           style="display: inline-block; padding: 14px 32px; background-color: #3182ce; color: #ffffff; text-decoration: none; border-radius: 6px; font-weight: 600; font-size: 16px;">
                                            Go to Dashboard
                                        </a>
                                    </td>
                                </tr>
                            </table>
                        </td>
                    </tr>

                    <!-- Footer -->
                    <tr>
                        <td style="padding: 20px 40px; border-top: 1px solid #e2e8f0; text-align: center;">
                            <p style="color: #718096; font-size: 14px; line-height: 1.6; margin: 0;">
                                © 2025 SaaS Blueprint. All rights reserved.
                            </p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>
`, name)
}

func renderVerificationEmail(token, baseURL string) string {
	verifyURL := fmt.Sprintf("%s/verify-email?token=%s", baseURL, token)

	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Verify Your Email</title>
</head>
<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;">
    <table width="100%%" cellpadding="0" cellspacing="0" style="background-color: #f7fafc; padding: 40px 20px;">
        <tr>
            <td align="center">
                <table width="600" cellpadding="0" cellspacing="0" style="background-color: #ffffff; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
                    <!-- Header -->
                    <tr>
                        <td style="padding: 40px 40px 20px 40px; text-align: center;">
                            <h1 style="margin: 0; color: #2d3748; font-size: 28px; font-weight: 700;">
                                Verify Your Email Address
                            </h1>
                        </td>
                    </tr>

                    <!-- Content -->
                    <tr>
                        <td style="padding: 0 40px 40px 40px;">
                            <p style="color: #4a5568; font-size: 16px; line-height: 1.6; margin: 0 0 20px 0;">
                                Thanks for signing up! Please verify your email address by clicking the button below.
                            </p>
                            <p style="color: #4a5568; font-size: 16px; line-height: 1.6; margin: 0 0 30px 0;">
                                This link will expire in 24 hours.
                            </p>

                            <!-- CTA Button -->
                            <table width="100%%" cellpadding="0" cellspacing="0">
                                <tr>
                                    <td align="center">
                                        <a href="%s"
                                           style="display: inline-block; padding: 14px 32px; background-color: #48bb78; color: #ffffff; text-decoration: none; border-radius: 6px; font-weight: 600; font-size: 16px;">
                                            Verify Email Address
                                        </a>
                                    </td>
                                </tr>
                            </table>

                            <p style="color: #718096; font-size: 14px; line-height: 1.6; margin: 30px 0 0 0;">
                                If the button doesn't work, copy and paste this link into your browser:
                            </p>
                            <p style="color: #3182ce; font-size: 14px; line-height: 1.6; margin: 10px 0 0 0; word-break: break-all;">
                                %s
                            </p>

                            <p style="color: #718096; font-size: 14px; line-height: 1.6; margin: 30px 0 0 0;">
                                If you didn't create an account, you can safely ignore this email.
                            </p>
                        </td>
                    </tr>

                    <!-- Footer -->
                    <tr>
                        <td style="padding: 20px 40px; border-top: 1px solid #e2e8f0; text-align: center;">
                            <p style="color: #718096; font-size: 14px; line-height: 1.6; margin: 0;">
                                © 2025 SaaS Blueprint. All rights reserved.
                            </p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>
`, verifyURL, verifyURL)
}

func renderPasswordResetEmail(token, baseURL string) string {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", baseURL, token)

	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Reset Your Password</title>
</head>
<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;">
    <table width="100%%" cellpadding="0" cellspacing="0" style="background-color: #f7fafc; padding: 40px 20px;">
        <tr>
            <td align="center">
                <table width="600" cellpadding="0" cellspacing="0" style="background-color: #ffffff; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
                    <!-- Header -->
                    <tr>
                        <td style="padding: 40px 40px 20px 40px; text-align: center;">
                            <h1 style="margin: 0; color: #2d3748; font-size: 28px; font-weight: 700;">
                                Reset Your Password
                            </h1>
                        </td>
                    </tr>

                    <!-- Content -->
                    <tr>
                        <td style="padding: 0 40px 40px 40px;">
                            <p style="color: #4a5568; font-size: 16px; line-height: 1.6; margin: 0 0 20px 0;">
                                We received a request to reset your password. Click the button below to create a new password.
                            </p>
                            <p style="color: #4a5568; font-size: 16px; line-height: 1.6; margin: 0 0 30px 0;">
                                This link will expire in 1 hour for security reasons.
                            </p>

                            <!-- CTA Button -->
                            <table width="100%%" cellpadding="0" cellspacing="0">
                                <tr>
                                    <td align="center">
                                        <a href="%s"
                                           style="display: inline-block; padding: 14px 32px; background-color: #f56565; color: #ffffff; text-decoration: none; border-radius: 6px; font-weight: 600; font-size: 16px;">
                                            Reset Password
                                        </a>
                                    </td>
                                </tr>
                            </table>

                            <p style="color: #718096; font-size: 14px; line-height: 1.6; margin: 30px 0 0 0;">
                                If the button doesn't work, copy and paste this link into your browser:
                            </p>
                            <p style="color: #3182ce; font-size: 14px; line-height: 1.6; margin: 10px 0 0 0; word-break: break-all;">
                                %s
                            </p>

                            <p style="color: #718096; font-size: 14px; line-height: 1.6; margin: 30px 0 0 0;">
                                If you didn't request a password reset, you can safely ignore this email. Your password will not be changed.
                            </p>
                        </td>
                    </tr>

                    <!-- Footer -->
                    <tr>
                        <td style="padding: 20px 40px; border-top: 1px solid #e2e8f0; text-align: center;">
                            <p style="color: #718096; font-size: 14px; line-height: 1.6; margin: 0;">
                                © 2025 SaaS Blueprint. All rights reserved.
                            </p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>
`, resetURL, resetURL)
}
