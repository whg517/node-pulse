# Story 1.1: Frontend Project Initialization

Status: done

## Story

As a 开发人员，
I want 初始化前端项目并配置基础依赖，
So that 后续可以开发前端功能。

## Acceptance Criteria

**Given** Pulse 前端项目不存在
**When** 开发人员执行初始化命令 `npm create vite@latest pulse-frontend -- --template react-ts`
**Then** 项目使用 React + TypeScript + Vite 创建成功
**And** Tailwind CSS 已正确安装和配置（tailwind.config.js）
**And** Apache ECharts 已安装
**And** 项目可以运行 `npm run dev` 启动开发服务器
**And** 基础项目结构已创建（src/components, src/pages, src/api, src/hooks）

## Tasks / Subtasks

- [x] Task 1: 创建 Vite + React + TypeScript 项目 (AC: #1)
  - [x] Subtask 1.1: 执行初始化命令创建项目
  - [x] Subtask 1.2: 验证项目结构正确
- [x] Task 2: 安装和配置 Tailwind CSS (AC: #2)
  - [x] Subtask 2.1: 安装 Tailwind CSS 及依赖
  - [x] Subtask 2.2: 初始化 Tailwind 配置文件
  - [x] Subtask 2.3: 配置 PostCSS 和 Tailwind 指令
- [x] Task 3: 安装 Apache ECharts (AC: #3)
  - [x] Subtask 3.1: 安装 echarts 包
- [x] Task 4: 创建基础项目结构 (AC: #4, #5)
  - [x] Subtask 4.1: 创建 src/components 目录
  - [x] Subtask 4.2: 创建 src/pages 目录
  - [x] Subtask 4.3: 创建 src/api 目录
  - [x] Subtask 4.4: 创建 src/hooks 目录
- [x] Task 5: 验证开发服务器启动 (AC: #5)
  - [x] Subtask 5.1: 运行 npm install
  - [x] Subtask 5.2: 运行 npm run dev
  - [x] Subtask 5.3: 验证开发服务器正常启动

## Dev Notes

### Architecture Requirements

**Frontend Technology Stack Decisions [Source: architecture.md#Starter Template Evaluation]:**

1. **Framework**: React + TypeScript + Vite
   - Rationale: 团队前端经验较弱，TypeScript 提供更好的类型安全和开发体验
   - Initialization command: `npm create vite@latest pulse-frontend -- --template react-ts`

2. **Styling**: Tailwind CSS
   - Rationale: Utility-first CSS 框架，轻量且不引入重型组件库
   - Installation:
     ```bash
     npm install -D tailwindcss postcss autoprefixer
     npx tailwindcss init -p
     ```

3. **Charts**: Apache ECharts
   - Rationale: 强大的图表库，支持复杂可视化需求
   - Installation: `npm install echarts`

### Project Structure Requirements

**Frontend Organization Pattern [Source: architecture.md#Frontend Architecture]:**

```
pulse-frontend/
├── public/              # 静态资源
├── src/
│   ├── components/       # React 组件
│   │   ├── common/      # 通用组件
│   │   ├── dashboard/   # 仪表盘专用
│   │   ├── nodes/       # 节点管理专用
│   │   └── alerts/      # 告警专用
│   ├── pages/           # 页面组件
│   ├── api/             # API 调用层
│   ├── hooks/           # 自定义 React Hooks
│   ├── types/           # TypeScript 类型定义
│   ├── utils/           # 工具函数
│   ├── App.tsx          # 根组件
│   └── main.tsx         # 应用入口
├── public/
├── tailwind.config.js   # Tailwind 配置
├── tsconfig.json        # TypeScript 配置
├── vite.config.ts       # Vite 配置
└── package.json
```

**Naming Conventions [Source: architecture.md#Code Naming Conventions]:**
- 组件文件使用 PascalCase: `ComponentName.tsx`
- 函数和变量使用 camelCase: `getUserData`, `createNode`
- 常量使用 UPPER_SNAKE_CASE: `MAX_RETRIES`, `DEFAULT_TIMEOUT`

### Implementation Sequence

**Critical Decision: This is the first implementation story** [Source: architecture.md#Starter Template Evaluation]

According to Architecture Section 334: "项目初始化使用此命令应该是第一个实现故事。"

**Follow this exact sequence:**

```bash
# Step 1: 创建项目
npm create vite@latest pulse-frontend -- --template react-ts

# Step 2: 进入项目目录
cd pulse-frontend

# Step 3: 安装依赖
npm install

# Step 4: 安装 Tailwind CSS
npm install -D tailwindcss postcss autoprefixer
npx tailwindcss init -p

# Step 5: 安装 Apache ECharts
npm install echarts

# Step 6: 创建基础目录结构
mkdir -p src/components/{common,dashboard,nodes,alerts}
mkdir -p src/pages
mkdir -p src/api
mkdir -p src/hooks
mkdir -p src/types
mkdir -p src/utils

# Step 7: 配置 Tailwind CSS
# 在 src/index.css 中添加 Tailwind 指令
# 配置 content 路径

# Step 8: 验证开发服务器
npm run dev
```

### Project Structure Notes

**No conflicts detected** - This is the first story, establishing the baseline structure.

**Component Organization Pattern [Source: architecture.md#Structure Patterns]:**
- 通用组件放在 `src/components/common/`
- 页面组件放在 `src/pages/`
- 组件按功能分组（dashboard/, nodes/, alerts/）

### References

- [Starter Template Decision] [Source: architecture.md#Starter Template Evaluation lines 210-334]
- [Frontend Architecture] [Source: architecture.md#Frontend Architecture lines 536-617]
- [Code Organization Pattern] [Source: architecture.md#Structure Patterns lines 1081-1119]
- [Naming Conventions] [Source: architecture.md#Naming Patterns lines 1056-1077]
- [Project Structure] [Source: architecture.md#Project Structure & Boundaries lines 1506-1565]

## Dev Agent Record

### Agent Model Used

claude-sonnet-4.5-20250929

### Debug Log References

None

### Completion Notes List

- Created Vite + React + TypeScript project using `npm create vite@latest pulse-frontend -- --template react-ts`
- Installed Tailwind CSS v3.4.0 (fixed from v4.1.18 to match architecture spec)
- Configured Tailwind CSS: tailwind.config.js, postcss.config.js with @tailwindcss/postcss plugin, src/index.css with directives
- Installed Apache ECharts
- Created directory structure: src/components/{common,dashboard,nodes,alerts}, src/pages, src/api, src/hooks, src/types, src/utils
- Configured test framework: Vitest + @testing-library/react with vitest.config.ts
- Created basic test: src/App.test.tsx
- Verified dev server starts successfully on http://localhost:5173/

### File List

pulse-frontend/.gitignore
pulse-frontend/eslint.config.js
pulse-frontend/index.html
pulse-frontend/package.json
pulse-frontend/package-lock.json
pulse-frontend/postcss.config.js
pulse-frontend/README.md
pulse-frontend/tailwind.config.js
pulse-frontend/tsconfig.app.json
pulse-frontend/tsconfig.json
pulse-frontend/tsconfig.node.json
pulse-frontend/vite.config.ts
pulse-frontend/vitest.config.ts
pulse-frontend/src/App.css
pulse-frontend/src/App.test.tsx
pulse-frontend/src/App.tsx
pulse-frontend/src/index.css
pulse-frontend/src/main.tsx
pulse-frontend/src/assets/
pulse-frontend/src/api/
pulse-frontend/src/components/alerts/
pulse-frontend/src/components/common/
pulse-frontend/src/components/dashboard/
pulse-frontend/src/components/nodes/
pulse-frontend/src/hooks/
pulse-frontend/src/pages/
pulse-frontend/src/types/
pulse-frontend/src/utils/
