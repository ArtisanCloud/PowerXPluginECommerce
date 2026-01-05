/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_MINIAPP_TENANT_UUID?: string;
  readonly VITE_MINIAPP_API_BASE?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}

