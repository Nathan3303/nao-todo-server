import { Todo } from '@nao-todo-server/models';
import { ObjectId, parseToBool } from '@nao-todo-server/utils';
import moment from 'moment';

const handleUserId = (userId?: string) => {
    return userId
        ? Todo.aggregate()
              .match({ userId: new ObjectId(userId) })
              .pipeline()
        : [];
};

const handleProjectId = (projectId?: string) => {
    return projectId
        ? Todo.aggregate()
              .match({ projectId: new ObjectId(projectId) })
              .pipeline()
        : [];
};

const handleTagId = (tagId?: string) => {
    return tagId
        ? Todo.aggregate()
              .match({ tags: { $in: [new ObjectId(tagId)] } })
              .pipeline()
        : [];
};

const handleId = (id: string) => {
    return id
        ? Todo.aggregate()
              .match({ _id: new ObjectId(id) })
              .pipeline()
        : [];
};

const handleName = (name?: string) => {
    return name
        ? Todo.aggregate()
              .match({ name: { $regex: name, $options: 'i' } })
              .pipeline()
        : [];
};

const handleState = (state?: string) => {
    if (!state) return [];
    const splitedState = state.split(',');
    return Todo.aggregate()
        .match({ state: { $in: splitedState } })
        .pipeline();
};

const handlePriority = (priority?: string) => {
    if (!priority) return [];
    const splitedPriority = priority.split(',');
    return Todo.aggregate()
        .match({ priority: { $in: splitedPriority } })
        .pipeline();
};

const handleIsFavorited = (isFavorited?: string) => {
    return isFavorited
        ? Todo.aggregate()
              .match({ isFavorited: parseToBool(isFavorited) })
              .pipeline()
        : [];
};

const handleIsDeleted = (isDeleted?: string) => {
    return isDeleted
        ? Todo.aggregate()
              .match({ isDeleted: parseToBool(isDeleted) })
              .pipeline()
        : [];
};

const handlePage = (page: string, limit: string) => {
    const _page = parseInt(page) || 1;
    const _limit = parseInt(limit) || 10;
    return Todo.aggregate()
        .skip((_page - 1) * _limit)
        .limit(_limit)
        .pipeline();
};

const handleSort = (sort?: string) => {
    if (!sort) {
        return Todo.aggregate().sort({ createdAt: 'desc' }).pipeline();
    }
    let [field, order] = sort.split(':');
    if (!field) return [];
    if (field === 'endAt') field = 'dueDate.endAt';
    else if (field === 'startAt') field = 'dueDate.startAt';
    order = order || 'asc';
    let _sortOption: Record<string, any> = {};
    _sortOption[field] = order;
    return Todo.aggregate().sort(_sortOption).pipeline();
};

const handleLookupProject = () => {
    return Todo.aggregate()
        .lookup({
            from: 'projects',
            localField: 'projectId',
            foreignField: '_id',
            as: '_project'
        })
        .unwind({ path: '$_project', preserveNullAndEmptyArrays: true })
        .pipeline();
};

const handleLookupTags = () => {
    return Todo.aggregate()
        .lookup({
            from: 'tags',
            localField: 'tags',
            foreignField: '_id',
            as: 'tagsInfo'
        })
        .pipeline();
};

const handleSelectFields = (fieldsOptions?: Record<string, any>) => {
    fieldsOptions = fieldsOptions || {
        _id: 0,
        id: { $toString: '$_id' },
        projectId: 1,
        project: { title: '$_project.title' },
        name: 1,
        description: 1,
        state: 1,
        priority: 1,
        dueDate: 1,
        isDone: 1,
        isFavorited: 1,
        isDeleted: 1,
        tags: 1,
        updatedAt: 1,
        createdAt: 1
    };
    return Todo.aggregate().project(fieldsOptions).pipeline();
};

const handleGroupByState = () => {
    return Todo.aggregate()
        .group({
            _id: '$state',
            total: { $sum: 1 },
            totalTodos: { $sum: 1 }
        })
        .pipeline();
};

const handleGroupByPriority = () => {
    return Todo.aggregate()
        .group({
            _id: '$priority',
            total: { $sum: 1 },
            totalTodos: { $sum: 1 }
        })
        .pipeline();
};

const handleCountTotal = () => {
    return Todo.aggregate().count('total').pipeline();
};

const handleRelativeDate = (relativeDate?: string) => {
    if (!relativeDate) return [];
    let agg = null;
    const now = moment();
    switch (relativeDate) {
        case 'today':
            agg = Todo.aggregate().match({
                'dueDate.endAt': {
                    $gte: now.startOf('day').toDate(),
                    $lte: now.endOf('day').toDate()
                }
            });
            break;
        case 'tomorrow':
            agg = Todo.aggregate().match({
                'dueDate.endAt': {
                    $gte: now.add(1, 'days').startOf('day').toDate(),
                    $lte: now.add(1, 'days').endOf('day').toDate()
                }
            });
            break;
        case 'week':
            agg = Todo.aggregate().match({
                'dueDate.endAt': {
                    $gte: now.startOf('isoWeek').toDate(),
                    $lte: now.endOf('isoWeek').toDate()
                }
            });
            break;
        default:
            agg = null;
    }
    return agg ? agg.sort({ 'dueDate.endAt': 1 }).pipeline() : [];
};

export default {
    handleUserId,
    handleProjectId,
    handleName,
    handlePriority,
    handleRelativeDate,
    handleState,
    handleTagId,
    handleId,
    handleIsFavorited,
    handleIsDeleted,
    handlePage,
    handleGroupByPriority,
    handleCountTotal,
    handleLookupProject,
    handleLookupTags,
    handleSelectFields,
    handleGroupByState,
    handleSort
};
