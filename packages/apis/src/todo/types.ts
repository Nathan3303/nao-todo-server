export type GetTodosQuery = {
    projectId?: string;
    name?: string;
    state?: string;
    priority?: string;
    isFavorited?: string;
    isDeleted?: string;
    relativeDate?: string;
    tagId?: string;
    sort?: string;
    page: string;
    limit: string;
};
