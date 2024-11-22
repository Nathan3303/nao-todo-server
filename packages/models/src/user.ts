import {
    getModelForClass,
    index,
    modelOptions,
    prop
} from '@typegoose/typegoose';

@index({ email: 1, password: 1 })
@modelOptions({ schemaOptions: { timestamps: true, collection: 'users' } })
class User {
    @prop({ unique: true })
    account: string;

    @prop({ required: true })
    password: string;

    @prop({ required: true, unique: true })
    email: string;

    @prop({ default: () => `User${Date.now()}` })
    nickname: string;

    @prop({
        default:
            'https://wpimg.wallstcn.com/f778738c-e4f8-4870-b634-56703b4acafe.gif?imageView2/1/w/80/h/80'
    })
    avatar: string;

    @prop({ required: true, default: 'user' })
    role: string;
}

export default getModelForClass(User);
export { User };
