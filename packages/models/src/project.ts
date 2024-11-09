import {
    getModelForClass,
    prop,
    modelOptions,
    index
} from '@typegoose/typegoose';

@modelOptions({ schemaOptions: { _id: false }, options: { allowMixed: 0 } })
class ProjectPreference {
    @prop({ default: 'table' })
    viewType: string;

    @prop({ default: { isDeleted: false } })
    filterInfo: object;

    @prop({ default: { field: 'createdAt', order: 'desc' } })
    sortInfo: object;

    @prop({
        default: {
            state: false,
            priority: true,
            project: true,
            description: true,
            endAt: true,
            createdAt: false,
            updatedAt: false
        }
    })
    columns: object;
}

@index({ userId: 1, title: 1 }, { unique: true })
@modelOptions({ schemaOptions: { timestamps: true, collection: 'projects' } })
class Project {
    @prop({ required: true })
    userId: string;

    @prop({ required: true })
    title: string;

    @prop({ default: null })
    description: string;

    @prop({ default: 0 })
    state: number;

    @prop({ default: false })
    isArchived: boolean;

    @prop({ default: false })
    isDeleted: boolean;

    @prop({ default: null })
    deletedAt: Date;

    @prop({ default: null })
    archivedAt: Date;

    @prop({ default: null })
    preference: ProjectPreference;
}

export default getModelForClass(Project);
export { Project };
