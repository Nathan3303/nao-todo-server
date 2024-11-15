import {
    getModelForClass,
    prop,
    modelOptions,
    index,
    mongoose
} from '@typegoose/typegoose';

@index({ userId: 1 }, { unique: true })
@modelOptions({ schemaOptions: { timestamps: true, collection: 'sessions' } })
class Session {
    @prop({ required: true, ref: 'User' })
    userId: mongoose.Types.ObjectId;

    @prop({ required: true })
    token: string;

    @prop({ required: true })
    expiresAt: Date;
}

export default getModelForClass(Session);
export { Session };
