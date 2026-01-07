import { miniAppRequest } from "./miniapp-client";

export type MiniAppCategoryNode = {
  id: string;
  parentId?: string | null;
  code?: string;
  displayName: string;
  aliasSlug?: string;
  path: string;
  level?: number;
  sortOrder?: number;
  status?: string;
  imageUrl?: string;
  isFeatured?: boolean;
  children?: MiniAppCategoryNode[];
};

export type MiniAppCategoryTreeResponse = {
  items: MiniAppCategoryNode[];
};

export async function miniAppGetCategoryTree() {
  return await miniAppRequest<MiniAppCategoryTreeResponse>({
    method: "GET",
    path: "/categories/tree",
  });
}

export async function miniAppGetCategoryTreeItems() {
  const data = await miniAppGetCategoryTree();
  return Array.isArray(data?.items) ? data.items : [];
}
