import { projectConfig } from './projectBase.config'

export type ProfileIconKey =
  | 'book-open'
  | 'globe'
  | 'message-square'
  | 'github'
  | 'mail'
  | 'external-link'

export interface ProfileChannelConfig {
  name: string
  description: string
  detail: string
  href?: string
  icon?: ProfileIconKey
}

export interface AuthorProfileConfig {
  name: string
  initial: string
  title: string
  bio: string
  location: string
  joinDate: string
  email: string
  website: string
  github: string
  skills: string[]
  channels: ProfileChannelConfig[]
}

export interface ProjectProfileActionConfig {
  label: string
  href: string
  icon: ProfileIconKey
}

export interface ProjectProfileConfig {
  name: string
  introBadge: string
  introText: string
  techStack: string[]
  description: string
  actions: ProjectProfileActionConfig[]
}

export interface RemoteAuthorSourceConfig {
  authorURL: string
  timeoutMs: number
}

export interface ProfilePageLocalConfig {
  remoteAuthor: RemoteAuthorSourceConfig
  defaultAuthor: AuthorProfileConfig
  project: ProjectProfileConfig
}

export const profilePageConfig: ProfilePageLocalConfig = {
  remoteAuthor: {
    // 留空时直接使用本地默认资料；需要远程作者页时再替换为真实地址。
    // https://static.lianxiang.local/profile/author.json
    // https://raw.githubusercontent.com/<user>/<repo>/main/author.json
    authorURL: '',
    timeoutMs: 1000,
  },
  defaultAuthor: {
    name: '链享浏览器使用教材',
    initial: '链',
    title: '从实例创建到自动化运行的操作指南',
    bio: '按浏览器实例、代理配置和自动化脚本三个步骤，快速完成链享浏览器的日常使用。所有资料均来自本地应用功能，不依赖外部作者页面。',
    location: '本地桌面应用',
    joinDate: '版本 1.3.0',
    email: '',
    website: '',
    github: '',
    skills: ['实例隔离', '代理绑定', '指纹配置', '自动化脚本', '内核管理', '数据备份'],
    channels: [
      {
        name: '第一章 实例管理',
        description: '创建独立浏览器环境',
        detail: '从实例列表开始，配置名称、指纹和启动参数。',
        icon: 'book-open',
      },
      {
        name: '第二章 代理配置',
        description: '导入节点并绑定实例',
        detail: '在代理池中导入、测速，再为实例选择代理。',
        icon: 'globe',
      },
      {
        name: '第三章 自动化',
        description: '运行脚本并查看结果',
        detail: '安装运行环境后，选择实例执行自动化任务。',
        icon: 'message-square',
      },
    ],
  },
  project: {
    name: projectConfig.name,
    introBadge: '使用教材',
    introText: '用于快速熟悉链享浏览器的核心工作流。',
    techStack: ['实例隔离', '代理池', '浏览器内核', '插件管理', '自动化', '数据备份'],
    description: '推荐顺序：先在实例列表创建环境，再到代理池配置并测速，最后按需设置指纹、插件和自动化脚本。遇到启动或网络问题时，优先查看内核管理和日志查看页面。',
    actions: [],
  },
}

export default profilePageConfig
