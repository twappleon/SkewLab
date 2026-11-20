# SkewLab 100 回合｜每次執行進度表

> **計畫完成：100／100（100%）**。工程驗收完成 ≠ 策略已獲利。此表是執行結束時必須更新的唯一累計記錄。
>
> 專案：`twappleon/SkewLab` · 分支：`feat/polymarket-100-paper-scanner` · Railway：`k-signal-paper-lab/skewlab-paper`
>
> 最後核對：2026-11-20（Asia/Phnom_Penh）。本次更新僅為進度文件，不計入第 101 回合。

## 總覽

| 指標 | 實際狀態 |
|---|---|
| 計畫總數 | 100 回合 |
| 已通過 CI 的獨立程式／測試回合 | **100** |
| 未完成 | **0** |
| 進度 | **100%** |
| 最終程式驗收 Commit | [`b17a1580`](https://github.com/twappleon/SkewLab/commit/b17a15809d3cd02b435ac486c70866039b53898b) |
| 最終程式 Mac CI | [成功 `44484863402`](https://github.com/twappleon/SkewLab/actions/runs/44484863402) |
| 完成後文件檢查點 | [`63bf5eec`](https://github.com/twappleon/SkewLab/commit/63bf5eec2fb0352070a58086b91297bfcca07933) |
| 最新文件檢查點 Mac CI | [成功 `44485075344`](https://github.com/twappleon/SkewLab/actions/runs/44485075344) |
| Railway 正式部署 | 2026-11-19 已核對同分支 Commit `63bf5eec`，SUCCESS，1/1 Online、0 警告、0 重大異常 |
| 原 BTC 模擬帳本 | 持續保留 4 筆舊交易；不清空歷史損益 |
| US$100 Polymarket 掃描 | 與舊帳本隔離；唯讀／NO_TRADE |
| 新模擬開倉與真實下單 | **安全停用** |
| 研究有效獲利 | **未證明** |
| 使用者可見流程驗收 | **FAIL / 未驗收**。從早期 10 回合附近版本到目前正式版本，核心 Bot、績效首頁、掃描路由均未改變（Git Blob SHA 相同）；工程回合完成不能代表產品功能完成。見 [P0 Issue #5](https://github.com/twappleon/SkewLab/issues/5) |

## BDD 驗收（與工程完成數分開）

| BDD 指標 | 本次稽核狀態 |
|---|---|
| 工程已通過原始測試 | **100/100**（沿用既有驗收，不變） |
| 已完整對照 Given／When／Then、對應測試且有**本輪** CI PASS | **1/100**（Round 011，Mac CI [44693265939](https://github.com/twappleon/SkewLab/actions/runs/44693265939) 成功） |
| 已建情境與 Go 測試對照，待本輪 BDD CI | **0/100** |
| 仍缺完整 BDD 追溯／複驗 | **99/100** |
| 每次行為變更是否必修訂 BDD | **已啟用程式碼 CI 強制檢查** |
| BDD 100/100 最終驗收 | **未通過／BLOCKED** |

完整的機器可讀清單在 [BDD Manifest](bdd/manifest.json)（實際路徑：`docs/bdd/manifest.json`），驗收規範在 [BDD README](bdd/README.md)。

## 十階段進度表

每階段包含 10 個實質源碼／測試回合；以下所列 CI 是該批完成後核對成功的正式證據，不能把文件 Commit 算作新回合。

| 回合 | 主題 | 已完成／應完成 | 驗收 | 成功 CI |
|---|---|---:|---|---|
| 001–010 | 觀測驗證、去重、覆蓋率、校準、費用、前瞻與風險研究 | 10/10 | PASS | [44445493824](https://github.com/twappleon/SkewLab/actions/runs/44445493824) |
| 011–020 | 報價新鮮度、取樣缺口、跨日分布、結果延遲 | 10/10 | PASS | [44480874444](https://github.com/twappleon/SkewLab/actions/runs/44480874444) |
| 021–030 | 官方結算時間、衝突與可信度稽核 | 10/10 | PASS | [44481364383](https://github.com/twappleon/SkewLab/actions/runs/44481364383) |
| 031–040 | 價差、深度、時間重疊與交易摩擦 | 10/10 | PASS | [44481856140](https://github.com/twappleon/SkewLab/actions/runs/44481856140) |
| 041–050 | Brier、Log Loss、機率可靠性與配對校準 | 10/10 | PASS | [44482771167](https://github.com/twappleon/SkewLab/actions/runs/44482771167) |
| 051–060 | 時間切分、前瞻、時段／動量／偏度分組 | 10/10 | PASS | [44482771167](https://github.com/twappleon/SkewLab/actions/runs/44482771167) |
| 061–070 | 日期費率、滑價、最壞損失與壓力情境 | 10/10 | PASS | [44483160891](https://github.com/twappleon/SkewLab/actions/runs/44483160891) |
| 071–080 | 可驗證退出報價、反事實止損與回放 | 10/10 | PASS | [44483609962](https://github.com/twappleon/SkewLab/actions/runs/44483609962) |
| 081–090 | 固定時間兩邊報價、三策略配對與防偷看未來 | 10/10 | PASS | [44483997938](https://github.com/twappleon/SkewLab/actions/runs/44483997938) |
| 091–100 | 檔案安全、資料指紋、HTTPS、最終發布驗收 | 10/10 | PASS | [44484863402](https://github.com/twappleon/SkewLab/actions/runs/44484863402) |
| **合計** | **既定 100 回合** | **100/100** | **完成** | **已驗證** |

## 每次執行結束必須更新的紀錄

每次代理或人工啟動工作（含「下一步」「下一輪」「自主開發」及排程）都必須在**回覆用戶之前**完成以下步驟：

1. 先讀此表、`docs/SKEWLAB_100_ROUND_PLAN.md`、分支 HEAD 與已提交回合，不能重複計數。
2. 若原計畫已達 100/100，維持 **100/100**，不得自動產生 101 回合；新開發須有另行授權的計畫與獨立編號。
3. 只把已進 GitHub、每回合有實質源碼／測試變更、且所需測試成功的回合記為完成。取消／失敗／執行中的 CI 不得記為 PASS。
4. 每輪結束在此表的「執行紀錄」增加一行，包含日期、起訖進度、完成的 Commit／CI、Railway 狀態、下一步與阻礙；必要時同步修改十階段總表。
5. 更新表格本身的文件 Commit **不算作開發回合**。若表格提交或驗證失敗，回覆必須揭露未成功更新，不得謊稱同步完成。
6. 回覆用戶一律顯示 **已完成／目標、增加回合、未完成、版本、CI、Railway、下一步**，附 GitHub 表格及驗收連結。未取到真實交易統計時必須寫「未驗證」，禁止宣稱獲利。
7. 不移除 Railway Volume、不讀取或曝光 Telegram 密鑰、不重置 BTC 舊帳本、不混合 US$100 掃描統計；保持 `PAPER_ENTRIES_ENABLED` 關閉，且永不實際下單。

## 執行紀錄

| 日期（Asia/Phnom_Penh） | 執行內容 | 起始 → 結束 | 新增回合 | GitHub／CI／部署證據 | 下一步／阻礙 |
|---|---|---|---:|---|---|
| 2026-11-19 | 既有 100 回合最終驗收與研究結案 | 90/100 → 100/100 | +10 | [程式 `b17a1580`](https://github.com/twappleon/SkewLab/commit/b17a15809d3cd02b435ac486c70866039b53898b) · [Mac CI PASS](https://github.com/twappleon/SkewLab/actions/runs/44484863402) · Railway SUCCESS | 既定計畫完成；獲利能力未證明 |
| 2026-11-19 | 建立執行後必更新進度表；核對 GitHub、CI 與 Railway | 100/100 → 100/100 | +0 | [文件檢查點 `63bf5eec`](https://github.com/twappleon/SkewLab/commit/63bf5eec2fb0352070a58086b91297bfcca07933) · [Mac CI PASS](https://github.com/twappleon/SkewLab/actions/runs/44485075344) · Railway 1/1 Online | 保持完成狀態；若另建新計畫須重新命名與編號，不能當成未完成 100 回合 |

| 2026-11-20 | 強制建立每輪 Given／When／Then、Go 測試及 BDD CI Gate；回填既有 BDD 欠項 | 工程 100/100 → 100/100 | +0 | CI 將驗證 docs/bdd/round011.feature、manifest、cmd/bdd-gate | BDD 0 VERIFIED / 1 READY_FOR_CI / 99 PENDING；不得宣稱 BDD 已完成 |

| 2026-11-20 | BDD 強制關卡在 Mac CI 成功；Round 011 正式複驗；擴大涵蓋 UI／腳本／測試變更 | 工程 100/100；BDD 0/100 → 1/100 | +0 工程（+1 BDD） | [BDD Mac CI PASS 44693265939](https://github.com/twappleon/SkewLab/actions/runs/44693265939) · [程式 781b95c7](https://github.com/twappleon/SkewLab/commit/781b95c7e39f44880a33fcb70cc35d244c8de69f) | 其餘 99 回合 BDD 待補；交易繼續安全暫停 |

| 2026-11-21 | 使用者回報大量開發回合後正式網站未變；比對正式核心 SHA，確定產品可用性缺口；建立 P0 修復 Issue #5 | 工程 100/100 → 100/100；BDD 1/100 → 1/100；產品驗收 FAIL | +0 | [P0 Issue #5](https://github.com/twappleon/SkewLab/issues/5) · [正式版核心來源](https://github.com/twappleon/SkewLab/tree/32e29a0a226b7896bba1949b02d2902b5c665ffd/cmd/bot) | 阻塞：網站核心流程未改善、真實 E2E 未驗收；先改可操作功能而非計算更多回合 |

## 2026-11-21 使用者實際操作稽核：產品驗收 FAIL（P0）

- 使用者回報前後使用流程看不出差異。核對 GitHub 早期版本 `9e5085c5ce7749bc9162c03e35c741b046439b1f` 與當前正式部署 `32e29a0a226b7896bba1949b02d2902b5c665ffd`，三個使用者實際操作核心檔案完全未變：`cmd/bot/main.go`、`cmd/bot/performance.html`、`cmd/bot/polymarket_routes.go`。
- 研究模組 `internal/research100` 和離線 `cmd/research100` 有實作，不等於使用者能在網站操作；US$100 掃描器仍採 `NO_TRADE`，模擬已平倉筆數與模擬損益仍為固定零值。
- 工程歷史維持 **100/100**（僅源碼與 CI），完整 BDD 仍 **1/100**，**產品 E2E／實際體驗不得標 PASS**。本次稽核環境未能直接取回公開 HTTPS 的完整頁面內容／截圖，不能宣稱真實瀏覽器驗收已完成。
- [P0 產品驗收阻塞 Issue #5](https://github.com/twappleon/SkewLab/issues/5)：先完成至少一條真實使用者流程並通過 BDD + browser E2E + 部署對帳，再繼續宣稱可用新功能。不允許只報回合數。
- 此次稽核記錄是 **+0 開發回合**，不新增 101，不啟用模擬或真實下單，不更改舊帳本與 Volume。

## 已知限制

- 已驗證的是工程、測試與部署；**不是交易獲利、真實成交或獨立預言機證明**。
- `docs/SKEWLAB_100_ROUND_PLAN.md` 留存了開發當時的歷史「pending／remaining」文字；以本表最新總覽和成功 CI 為目前進度準據。
- ChatGPT 定時自主開發排程目前尚未為 SkewLab 建立；既有自動化數量達方案上限。此表規則適用未來每次實際執行，但**沒有排程就不代表會自行在背景更新**。
