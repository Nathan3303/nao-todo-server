import {
    getModelForClass,
    prop,
    modelOptions,
    index
} from '@typegoose/typegoose';

@modelOptions({ schemaOptions: { _id: false } })
class TodoDueDate {
    @prop({ default: new Date() })
    startAt: Date;

    @prop({ default: null })
    endAt: Date;
}

@index({ userId: 1, projectId: 1 })
@index({ userId: 1, projectId: 1, name: 1, state: 1, priority: 1 })
@modelOptions({
    schemaOptions: { timestamps: true, collection: 'todos' },
    options: { allowMixed: 0 }
})
class Todo {
    @prop({ required: true })
    userId: string;

    @prop({ required: true })
    projectId: string;

    @prop({ required: true })
    name: string;

    @prop({ default: '' })
    description: string;

    @prop({ default: 'todo', enum: ['todo', 'doing', 'done'] })
    state: string;

    @prop({ default: 'low', enum: ['low', 'medium', 'high', 'urgent'] })
    priority: string;

    @prop({ default: [] })
    tags: string[];

    @prop({ default: { startAt: null, endAt: null } })
    dueDate: TodoDueDate;

    @prop({ default: false })
    isDeleted: boolean;

    @prop({ default: false })
    isArchived: boolean;

    @prop({ default: false })
    isFavorited: boolean;
}

export default getModelForClass(Todo);
export { Todo };
