import type { DashboardMessages } from "@yueli/ui/dashboard/pattern";

export const resourceDashboardMessages: DashboardMessages = {
  metrics: "关键指标",
  pending: {
    title: "待处理",
    description: "优先处理会阻塞发布或协作的工作。",
  },
  recent: {
    title: "最近工作",
    description: "继续处理最近更新的内容。",
  },
  health: {
    title: "当前站点健康",
    description: "只显示会影响当前站点任务的状态。",
  },
  quickActions: {
    title: "快捷动作",
    description: "进入最常使用的下一步。",
  },
};
