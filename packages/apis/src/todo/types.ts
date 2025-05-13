export type GetTodosQuery = {
    projectId?: string;
    name?: string;
    state?: string;
    priority?: string;
    isFavorited?: string;
    isDeleted?: string
    isGivenUp?: string;
    relativeDate?: string;
    tagId?: string;
    sort?: string;
    page: string;
    limit: string;
};

export type TodoDueDate = {
    startAt: null | string | Date,
    endAt: null | string | Date
}
