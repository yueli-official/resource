const absoluteDate = new Intl.DateTimeFormat("zh-CN", {
  year: "numeric",
  month: "long",
  day: "numeric",
});

/** 固定中文绝对日期；依赖时区的结果只应在客户端渲染。 */
export function abs(iso?: string): string {
  if (!iso) return "";
  const value = new Date(iso);
  return Number.isNaN(value.valueOf()) ? "" : absoluteDate.format(value);
}
