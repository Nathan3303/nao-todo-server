export type GetTodosQuery = {
    projectId?: string;
    name?: string;
    state?: string;
    priority?: string;
    isFavorited?: boolean;
    isDeleted?: boolean;
    relativeDate?: string;
    sort?: string;
    page: string;
    limit: string;
};
