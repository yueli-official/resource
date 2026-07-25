interface ResourceMeResponse {
  isAdministrator: boolean;
  roles: string[];
  capabilities: string[];
}

export function useResourceMe() {
  const { call } = useApi();
  const { loggedIn } = useAuth();
  const { data, pending, refresh } = useAsyncData(
    "resource-me",
    () => loggedIn.value
      ? call<ResourceMeResponse>("/api/v1/me").catch(() => null)
      : Promise.resolve(null),
    { server: false, watch: [loggedIn] },
  );
  return {
    me: data,
    isAdministrator: computed(() => !!data.value?.isAdministrator),
    can: (capability: string) => data.value?.capabilities?.includes(capability) ?? false,
    canManage: computed(() => (data.value?.capabilities?.length ?? 0) > 0),
    roles: computed(() => data.value?.roles ?? []),
    pending,
    refreshMe: refresh,
  };
}
