/****************************************************************************
** Meta object code from reading C++ file 'infomodel.h'
**
** Created by: The Qt Meta Object Compiler version 67 (Qt 5.5.0)
**
** WARNING! All changes made in this file will be lost!
*****************************************************************************/

#include "infomodel.h"
#include <QtCore/qbytearray.h>
#include <QtCore/qmetatype.h>
#if !defined(Q_MOC_OUTPUT_REVISION)
#error "The header file 'infomodel.h' doesn't include <QObject>."
#elif Q_MOC_OUTPUT_REVISION != 67
#error "This file was generated using the moc from 5.5.0. It"
#error "cannot be used with the include files from this version of Qt."
#error "(The moc has changed too much.)"
#endif

QT_BEGIN_MOC_NAMESPACE
struct qt_meta_stringdata_ComicType_t {
    QByteArrayData data[9];
    char stringdata0[66];
};
#define QT_MOC_LITERAL(idx, ofs, len) \
    Q_STATIC_BYTE_ARRAY_DATA_HEADER_INITIALIZER_WITH_OFFSET(len, \
    qptrdiff(offsetof(qt_meta_stringdata_ComicType_t, stringdata0) + ofs \
        - idx * sizeof(QByteArrayData)) \
    )
static const qt_meta_stringdata_ComicType_t qt_meta_stringdata_ComicType = {
    {
QT_MOC_LITERAL(0, 0, 9), // "ComicType"
QT_MOC_LITERAL(1, 10, 4), // "Enum"
QT_MOC_LITERAL(2, 15, 7), // "Invalid"
QT_MOC_LITERAL(3, 23, 5), // "Manga"
QT_MOC_LITERAL(4, 29, 6), // "Manhwa"
QT_MOC_LITERAL(5, 36, 6), // "Manhua"
QT_MOC_LITERAL(6, 43, 7), // "Western"
QT_MOC_LITERAL(7, 51, 8), // "Webcomic"
QT_MOC_LITERAL(8, 60, 5) // "Other"

    },
    "ComicType\0Enum\0Invalid\0Manga\0Manhwa\0"
    "Manhua\0Western\0Webcomic\0Other"
};
#undef QT_MOC_LITERAL

static const uint qt_meta_data_ComicType[] = {

 // content:
       7,       // revision
       0,       // classname
       0,    0, // classinfo
       0,    0, // methods
       0,    0, // properties
       1,   14, // enums/sets
       0,    0, // constructors
       0,       // flags
       0,       // signalCount

 // enums: name, flags, count, data
       1, 0x0,    7,   18,

 // enum data: key, value
       2, uint(ComicType::Invalid),
       3, uint(ComicType::Manga),
       4, uint(ComicType::Manhwa),
       5, uint(ComicType::Manhua),
       6, uint(ComicType::Western),
       7, uint(ComicType::Webcomic),
       8, uint(ComicType::Other),

       0        // eod
};

void ComicType::qt_static_metacall(QObject *_o, QMetaObject::Call _c, int _id, void **_a)
{
    Q_UNUSED(_o);
    Q_UNUSED(_id);
    Q_UNUSED(_c);
    Q_UNUSED(_a);
}

const QMetaObject ComicType::staticMetaObject = {
    { &QObject::staticMetaObject, qt_meta_stringdata_ComicType.data,
      qt_meta_data_ComicType,  qt_static_metacall, Q_NULLPTR, Q_NULLPTR}
};


const QMetaObject *ComicType::metaObject() const
{
    return QObject::d_ptr->metaObject ? QObject::d_ptr->dynamicMetaObject() : &staticMetaObject;
}

void *ComicType::qt_metacast(const char *_clname)
{
    if (!_clname) return Q_NULLPTR;
    if (!strcmp(_clname, qt_meta_stringdata_ComicType.stringdata0))
        return static_cast<void*>(const_cast< ComicType*>(this));
    return QObject::qt_metacast(_clname);
}

int ComicType::qt_metacall(QMetaObject::Call _c, int _id, void **_a)
{
    _id = QObject::qt_metacall(_c, _id, _a);
    if (_id < 0)
        return _id;
    return _id;
}
struct qt_meta_stringdata_ComicStatus_t {
    QByteArrayData data[7];
    char stringdata0[64];
};
#define QT_MOC_LITERAL(idx, ofs, len) \
    Q_STATIC_BYTE_ARRAY_DATA_HEADER_INITIALIZER_WITH_OFFSET(len, \
    qptrdiff(offsetof(qt_meta_stringdata_ComicStatus_t, stringdata0) + ofs \
        - idx * sizeof(QByteArrayData)) \
    )
static const qt_meta_stringdata_ComicStatus_t qt_meta_stringdata_ComicStatus = {
    {
QT_MOC_LITERAL(0, 0, 11), // "ComicStatus"
QT_MOC_LITERAL(1, 12, 4), // "Enum"
QT_MOC_LITERAL(2, 17, 7), // "Invalid"
QT_MOC_LITERAL(3, 25, 8), // "Complete"
QT_MOC_LITERAL(4, 34, 7), // "Ongoing"
QT_MOC_LITERAL(5, 42, 8), // "OnHiatus"
QT_MOC_LITERAL(6, 51, 12) // "Discontinued"

    },
    "ComicStatus\0Enum\0Invalid\0Complete\0"
    "Ongoing\0OnHiatus\0Discontinued"
};
#undef QT_MOC_LITERAL

static const uint qt_meta_data_ComicStatus[] = {

 // content:
       7,       // revision
       0,       // classname
       0,    0, // classinfo
       0,    0, // methods
       0,    0, // properties
       1,   14, // enums/sets
       0,    0, // constructors
       0,       // flags
       0,       // signalCount

 // enums: name, flags, count, data
       1, 0x0,    5,   18,

 // enum data: key, value
       2, uint(ComicStatus::Invalid),
       3, uint(ComicStatus::Complete),
       4, uint(ComicStatus::Ongoing),
       5, uint(ComicStatus::OnHiatus),
       6, uint(ComicStatus::Discontinued),

       0        // eod
};

void ComicStatus::qt_static_metacall(QObject *_o, QMetaObject::Call _c, int _id, void **_a)
{
    Q_UNUSED(_o);
    Q_UNUSED(_id);
    Q_UNUSED(_c);
    Q_UNUSED(_a);
}

const QMetaObject ComicStatus::staticMetaObject = {
    { &QObject::staticMetaObject, qt_meta_stringdata_ComicStatus.data,
      qt_meta_data_ComicStatus,  qt_static_metacall, Q_NULLPTR, Q_NULLPTR}
};


const QMetaObject *ComicStatus::metaObject() const
{
    return QObject::d_ptr->metaObject ? QObject::d_ptr->dynamicMetaObject() : &staticMetaObject;
}

void *ComicStatus::qt_metacast(const char *_clname)
{
    if (!_clname) return Q_NULLPTR;
    if (!strcmp(_clname, qt_meta_stringdata_ComicStatus.stringdata0))
        return static_cast<void*>(const_cast< ComicStatus*>(this));
    return QObject::qt_metacast(_clname);
}

int ComicStatus::qt_metacall(QMetaObject::Call _c, int _id, void **_a)
{
    _id = QObject::qt_metacall(_c, _id, _a);
    if (_id < 0)
        return _id;
    return _id;
}
struct qt_meta_stringdata_ScanlationStatus_t {
    QByteArrayData data[8];
    char stringdata0[91];
};
#define QT_MOC_LITERAL(idx, ofs, len) \
    Q_STATIC_BYTE_ARRAY_DATA_HEADER_INITIALIZER_WITH_OFFSET(len, \
    qptrdiff(offsetof(qt_meta_stringdata_ScanlationStatus_t, stringdata0) + ofs \
        - idx * sizeof(QByteArrayData)) \
    )
static const qt_meta_stringdata_ScanlationStatus_t qt_meta_stringdata_ScanlationStatus = {
    {
QT_MOC_LITERAL(0, 0, 16), // "ScanlationStatus"
QT_MOC_LITERAL(1, 17, 4), // "Enum"
QT_MOC_LITERAL(2, 22, 7), // "Invalid"
QT_MOC_LITERAL(3, 30, 8), // "Complete"
QT_MOC_LITERAL(4, 39, 7), // "Ongoing"
QT_MOC_LITERAL(5, 47, 8), // "OnHiatus"
QT_MOC_LITERAL(6, 56, 7), // "Dropped"
QT_MOC_LITERAL(7, 64, 26) // "InDesperateNeedOfMoreStaff"

    },
    "ScanlationStatus\0Enum\0Invalid\0Complete\0"
    "Ongoing\0OnHiatus\0Dropped\0"
    "InDesperateNeedOfMoreStaff"
};
#undef QT_MOC_LITERAL

static const uint qt_meta_data_ScanlationStatus[] = {

 // content:
       7,       // revision
       0,       // classname
       0,    0, // classinfo
       0,    0, // methods
       0,    0, // properties
       1,   14, // enums/sets
       0,    0, // constructors
       0,       // flags
       0,       // signalCount

 // enums: name, flags, count, data
       1, 0x0,    6,   18,

 // enum data: key, value
       2, uint(ScanlationStatus::Invalid),
       3, uint(ScanlationStatus::Complete),
       4, uint(ScanlationStatus::Ongoing),
       5, uint(ScanlationStatus::OnHiatus),
       6, uint(ScanlationStatus::Dropped),
       7, uint(ScanlationStatus::InDesperateNeedOfMoreStaff),

       0        // eod
};

void ScanlationStatus::qt_static_metacall(QObject *_o, QMetaObject::Call _c, int _id, void **_a)
{
    Q_UNUSED(_o);
    Q_UNUSED(_id);
    Q_UNUSED(_c);
    Q_UNUSED(_a);
}

const QMetaObject ScanlationStatus::staticMetaObject = {
    { &QObject::staticMetaObject, qt_meta_stringdata_ScanlationStatus.data,
      qt_meta_data_ScanlationStatus,  qt_static_metacall, Q_NULLPTR, Q_NULLPTR}
};


const QMetaObject *ScanlationStatus::metaObject() const
{
    return QObject::d_ptr->metaObject ? QObject::d_ptr->dynamicMetaObject() : &staticMetaObject;
}

void *ScanlationStatus::qt_metacast(const char *_clname)
{
    if (!_clname) return Q_NULLPTR;
    if (!strcmp(_clname, qt_meta_stringdata_ScanlationStatus.stringdata0))
        return static_cast<void*>(const_cast< ScanlationStatus*>(this));
    return QObject::qt_metacast(_clname);
}

int ScanlationStatus::qt_metacall(QMetaObject::Call _c, int _id, void **_a)
{
    _id = QObject::qt_metacall(_c, _id, _a);
    if (_id < 0)
        return _id;
    return _id;
}
struct qt_meta_stringdata_ComicInfoModel_t {
    QByteArrayData data[6];
    char stringdata0[43];
};
#define QT_MOC_LITERAL(idx, ofs, len) \
    Q_STATIC_BYTE_ARRAY_DATA_HEADER_INITIALIZER_WITH_OFFSET(len, \
    qptrdiff(offsetof(qt_meta_stringdata_ComicInfoModel_t, stringdata0) + ofs \
        - idx * sizeof(QByteArrayData)) \
    )
static const qt_meta_stringdata_ComicInfoModel_t qt_meta_stringdata_ComicInfoModel = {
    {
QT_MOC_LITERAL(0, 0, 14), // "ComicInfoModel"
QT_MOC_LITERAL(1, 15, 6), // "qmlGet"
QT_MOC_LITERAL(2, 22, 0), // ""
QT_MOC_LITERAL(3, 23, 3), // "row"
QT_MOC_LITERAL(4, 27, 6), // "column"
QT_MOC_LITERAL(5, 34, 8) // "roleName"

    },
    "ComicInfoModel\0qmlGet\0\0row\0column\0"
    "roleName"
};
#undef QT_MOC_LITERAL

static const uint qt_meta_data_ComicInfoModel[] = {

 // content:
       7,       // revision
       0,       // classname
       0,    0, // classinfo
       1,   14, // methods
       0,    0, // properties
       0,    0, // enums/sets
       0,    0, // constructors
       0,       // flags
       0,       // signalCount

 // methods: name, argc, parameters, tag, flags
       1,    3,   19,    2, 0x02 /* Public */,

 // methods: parameters
    QMetaType::QVariant, QMetaType::Int, QMetaType::Int, QMetaType::QString,    3,    4,    5,

       0        // eod
};

void ComicInfoModel::qt_static_metacall(QObject *_o, QMetaObject::Call _c, int _id, void **_a)
{
    if (_c == QMetaObject::InvokeMetaMethod) {
        ComicInfoModel *_t = static_cast<ComicInfoModel *>(_o);
        Q_UNUSED(_t)
        switch (_id) {
        case 0: { QVariant _r = _t->qmlGet((*reinterpret_cast< int(*)>(_a[1])),(*reinterpret_cast< int(*)>(_a[2])),(*reinterpret_cast< const QString(*)>(_a[3])));
            if (_a[0]) *reinterpret_cast< QVariant*>(_a[0]) = _r; }  break;
        default: ;
        }
    }
}

const QMetaObject ComicInfoModel::staticMetaObject = {
    { &QAbstractTableModel::staticMetaObject, qt_meta_stringdata_ComicInfoModel.data,
      qt_meta_data_ComicInfoModel,  qt_static_metacall, Q_NULLPTR, Q_NULLPTR}
};


const QMetaObject *ComicInfoModel::metaObject() const
{
    return QObject::d_ptr->metaObject ? QObject::d_ptr->dynamicMetaObject() : &staticMetaObject;
}

void *ComicInfoModel::qt_metacast(const char *_clname)
{
    if (!_clname) return Q_NULLPTR;
    if (!strcmp(_clname, qt_meta_stringdata_ComicInfoModel.stringdata0))
        return static_cast<void*>(const_cast< ComicInfoModel*>(this));
    return QAbstractTableModel::qt_metacast(_clname);
}

int ComicInfoModel::qt_metacall(QMetaObject::Call _c, int _id, void **_a)
{
    _id = QAbstractTableModel::qt_metacall(_c, _id, _a);
    if (_id < 0)
        return _id;
    if (_c == QMetaObject::InvokeMetaMethod) {
        if (_id < 1)
            qt_static_metacall(this, _c, _id, _a);
        _id -= 1;
    } else if (_c == QMetaObject::RegisterMethodArgumentMetaType) {
        if (_id < 1)
            *reinterpret_cast<int*>(_a[0]) = -1;
        _id -= 1;
    }
    return _id;
}
QT_END_MOC_NAMESPACE
