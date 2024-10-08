import { createElement, useEffect, useState } from "react";

import Image from "next/image";
import Link from "next/link";
import router from "next/router";

import {
    ArrowRightOnRectangleIcon,
    Bars2Icon,
    CalendarDaysIcon,
    ChevronDownIcon,
    CogIcon,
    TicketIcon,
} from "@heroicons/react/24/outline";
import {
    Button,
    Collapse,
    IconButton,
    Menu,
    MenuHandler,
    MenuItem,
    MenuList,
    Navbar,
    Typography,
} from "@material-tailwind/react";

import logOut from "@/util/logOut";

import { useFirebaseAuth } from "@/components/FirebaseAuthContext";

import DefaultAvatar from "@/assets/default-avatar.jpg";

const profileMenuItems = [
    {
        label: "Sign Out",
        icon: ArrowRightOnRectangleIcon,
        action: () => logOut(),
    },
];

const adminProfileMenuItems = [
    {
        label: "Open Admin Portal",
        icon: CogIcon,
        action: () => router.push("/admin"),
    },
];

function ProfileMenu() {
    const [isMenuOpen, setIsMenuOpen] = useState(false);
    const { user } = useFirebaseAuth();
    const [isAdmin, setIsAdmin] = useState(false);

    const closeMenu = () => setIsMenuOpen(false);

    useEffect(() => {
        (async () => {
            user?.getIdTokenResult().then((res) => {
                setIsAdmin(!!res.claims["admin"]);
            });
        })();
    });

    return (
        <Menu
            open={isMenuOpen}
            handler={setIsMenuOpen}
            placement="bottom-end"
        >
            <MenuHandler>
                <Button
                    variant="text"
                    color="blue-gray"
                    className="flex items-center gap-1 rounded-full py-0.5 pr-2 pl-0.5 md:ml-auto"
                >
                    <Image
                        src={user?.photoURL ?? DefaultAvatar}
                        width={32}
                        height={32}
                        alt="profile picture"
                        className="border border-gray-900 p-0.5 rounded-full object-cover object-center"
                        unoptimized
                    />

                    <ChevronDownIcon
                        strokeWidth={2.5}
                        className={`h-3 w-3 transition-transform ${isMenuOpen ? "rotate-180" : ""}`}
                    />
                </Button>
            </MenuHandler>
            <MenuList className="p-1">
                {isAdmin &&
                    adminProfileMenuItems.map(({ label, icon, action }, key) => {
                        const isLastItem = key === profileMenuItems.length - 1;
                        return (
                            <MenuItem
                                key={label}
                                onClick={() => {
                                    action();
                                    closeMenu();
                                }}
                                className="flex items-center gap-2 rounded"
                            >
                                {createElement(icon, {
                                    className: "h-4 w-4",
                                    strokeWidth: 2,
                                })}
                                <Typography
                                    as="span"
                                    variant="small"
                                    className="font-normal"
                                    color="inherit"
                                >
                                    {label}
                                </Typography>
                            </MenuItem>
                        );
                    })}

                {profileMenuItems.map(({ label, icon, action }, key) => {
                    const isLastItem = key === profileMenuItems.length - 1;
                    return (
                        <MenuItem
                            key={label}
                            onClick={() => {
                                action();
                                closeMenu();
                            }}
                            className={`flex items-center gap-2 rounded ${
                                isLastItem ? "hover:bg-red-500/10 focus:bg-red-500/10 active:bg-red-500/10" : ""
                            }`}
                        >
                            {createElement(icon, {
                                className: `h-4 w-4 ${isLastItem ? "text-red-500" : ""}`,
                                strokeWidth: 2,
                            })}
                            <Typography
                                as="span"
                                variant="small"
                                className="font-normal"
                                color={isLastItem ? "red" : "inherit"}
                            >
                                {label}
                            </Typography>
                        </MenuItem>
                    );
                })}
            </MenuList>
        </Menu>
    );
}

const navListItems = [
    {
        label: "Events",
        icon: CalendarDaysIcon,
        link: "/events",
    },
    {
        label: "Tickets",
        icon: TicketIcon,
        link: "/tickets",
    },
];

function NavList() {
    return (
        <ul className="my-2 flex flex-col gap-2 md:mb-0 md:mt-0 md:flex-row md:items-center">
            {navListItems.map(({ label, icon, link }, key) => (
                <Link
                    href={link}
                    key={key}
                >
                    <Typography
                        variant="small"
                        color="blue-gray"
                        className="font-normal"
                    >
                        <MenuItem className="flex items-center gap-2 md:rounded-full">
                            {createElement(icon, {
                                className: "h-[18px] w-[18px]",
                            })}{" "}
                            {label}
                        </MenuItem>
                    </Typography>
                </Link>
            ))}
        </ul>
    );
}

export function ComplexNavbar() {
    const [isNavOpen, setIsNavOpen] = useState(false);

    const toggleIsNavOpen = () => setIsNavOpen((cur) => !cur);

    useEffect(() => {
        window.addEventListener("resize", () => window.innerWidth >= 768 && setIsNavOpen(false));
    }, []);

    return (
        <Navbar className="md:mx-4 md:mt-4 p-2 rounded-none md:rounded-full md:pl-6 w-auto transition-all duration-150 max-w-none">
            <div className="relative mx-auto flex items-center text-blue-gray-900">
                <Link href="/events">
                    <Typography className="mr-4 ml-2 cursor-pointer font-medium text-xl">FraserTickets</Typography>
                </Link>
                <div className="absolute top-2/4 left-2/4 hidden -translate-x-2/4 -translate-y-2/4 md:block">
                    <NavList />
                </div>
                <IconButton
                    size="sm"
                    color="blue-gray"
                    variant="text"
                    onClick={toggleIsNavOpen}
                    className="ml-auto mr-2 md:hidden"
                >
                    <Bars2Icon className="h-6 w-6" />
                </IconButton>
                <ProfileMenu />
            </div>
            <Collapse
                open={isNavOpen}
                className="h-full"
            >
                <NavList />
            </Collapse>
        </Navbar>
    );
}
