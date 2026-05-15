// thumbnail-worker 由 Go 后端按需启动，用本地 Chromium 生成网页缩略图截图。
import { chromium } from 'playwright'

const [, , rawUrl, outputPath] = process.argv

if (!rawUrl || !outputPath) {
  console.error('usage: thumbnail-worker <url> <output>')
  process.exit(2)
}

let browser
try {
  browser = await chromium.launch({ headless: true })
  const page = await browser.newPage({
    viewport: { width: 1200, height: 630 },
    deviceScaleFactor: 1
  })
  await page.goto(rawUrl, { waitUntil: 'networkidle', timeout: 20000 })
  await page.screenshot({ path: outputPath, type: 'png', fullPage: false })
} catch (error) {
  console.error(error instanceof Error ? error.message : String(error))
  process.exitCode = 1
} finally {
  await browser?.close().catch(() => undefined)
}
