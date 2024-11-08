import {
    getModelForClass,
    prop,
    modelOptions,
    index
} from '@typegoose/typegoose';

@index({ userId: 1 }, { unique: true })
@modelOptions({ schemaOptions: { timestamps: true, collection: 'sessions' } })
class Session {
    @prop({ required: true })
    userId: string;

    @prop({ required: true })
    token: string;

    @prop({ required: true })
    expiresAt: Date;
}

export default getModelForClass(Session);
export { Session };
